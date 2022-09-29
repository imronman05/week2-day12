package main

import (
	"context"
	"day7/connection"
	"day7/middleware"
	"fmt"
	"html/template"
	"log"
	"strings"

	"math"
	"net/http"
	"strconv"

	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	route := mux.NewRouter()

	connection.ConnectionProject()

	route.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))))

	route.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads/"))))

	route.PathPrefix("/node_modules/").Handler(http.StripPrefix("/node_modules/", http.FileServer(http.Dir("./node_modules"))))

	route.HandleFunc("/", home).Methods("GET")
	route.HandleFunc("/form-project", formProject).Methods("GET")
	route.HandleFunc("/add-project", middleware.ImageUpload(addProject)).Methods("POST")
	route.HandleFunc("/detail-project/{id}", detailProject).Methods("GET")
	route.HandleFunc("/contact", contact).Methods("GET")
	route.HandleFunc("/delete-project/{id}", deleteProject).Methods("GET")
	route.HandleFunc("/edit-project/{id}", editProject).Methods("GET")
	route.HandleFunc("/update-project/{id}", middleware.EditImage(updateProject)).Methods("POST")
	route.HandleFunc("/login", login).Methods("GET")
	route.HandleFunc("/form-login", formLogin).Methods("POST")
	route.HandleFunc("/register", register).Methods("GET")
	route.HandleFunc("/form-register", formRegister).Methods("POST")
	route.HandleFunc("/logout", logout).Methods("GET")

	fmt.Println("server on")
	http.ListenAndServe("localhost:5000", route)
}

type sessionData struct {
	Login     bool
	UserName  string
	FlashData string
}

var Data = sessionData{}

type Blog struct {
	Id              int
	NameProject     string
	StarDate        time.Time
	Format_starDate string
	Edit_starDate   string
	EndDate         time.Time
	Format_endDate  string
	Edit_endDate    string
	Duration        string
	Message         string
	Tech            []string
	Image           string
	UserId          string
}

type User struct {
	Id       int
	Name     string
	Email    string
	Password string
}

func addProject(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()

	if err != nil {
		log.Fatal(err)
	}

	var nameProject = r.PostForm.Get("projectName")
	var startDate = r.PostForm.Get("startDate")
	var endDate = r.PostForm.Get("endDate")
	var desc = r.PostForm.Get("Description")

	dataContext := r.Context().Value("dataFile")
	image := dataContext.(string)

	var tech []string
	for key, values := range r.Form {
		for _, value := range values {
			if key == "technologies" {
				tech = append(tech, value)
			}
		}
	}

	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")

	userId := session.Values["Id"].(int)
	// fmt.Println(userId)

	_, err = connection.Conn.Exec(context.Background(), "INSERT INTO tb_projects(name,start_date,end_date,description,technologies,image,user_id) VALUES ($1,$2,$3,$4,$5,$6,$7)", nameProject, startDate, endDate, desc, tech, image, userId)

	if err != nil {
		w.Write([]byte("Error baris 78 " + err.Error()))
	}
	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var tmpl, err = template.ParseFiles("view/index.html")

	if err != nil {
		w.Write([]byte("error 119: " + err.Error()))
		return
	}

	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")

	if session.Values["Login"] != true {
		Data.Login = false
	} else {
		Data.Login = session.Values["Login"].(bool)
		Data.UserName = session.Values["Nama"].(string)
	}

	flasMessege := session.Flashes("dataSession")

	var flashes []string
	if len(flasMessege) > 0 {
		session.Save(r, w)
		for _, f1 := range flasMessege {
			flashes = append(flashes, f1.(string))
		}
	}

	Data.FlashData = strings.Join(flashes, "")

	if session.Values["Login"] != true {
		data, _ := connection.Conn.Query(context.Background(), "SELECT id, name, start_date, end_date, description, technologies,image FROM tb_projects")

		var result []Blog

		for data.Next() {
			var each = Blog{}

			err := data.Scan(&each.Id, &each.NameProject, &each.StarDate, &each.EndDate, &each.Message, &each.Tech, &each.Image)

			if err != nil {
				fmt.Println("Error baris 126 " + err.Error())
				return
			}

			hs := each.EndDate.Sub(each.StarDate).Hours()
			day, _ := math.Modf(hs / 24)
			bulan := int64(day / 30)
			tahun := int64(day / 365)

			if tahun > 0 {
				each.Duration = strconv.FormatInt(tahun, 10) + " Year"
			} else if bulan > 0 {
				each.Duration = strconv.FormatInt(bulan, 10) + " Month"
			} else {
				each.Duration = fmt.Sprintf("%.0f", day) + " Day"
			}
			result = append(result, each)
		}
		// fmt.Println(result)
		response := map[string]interface{}{
			"DataSession": Data,
			"Blogs":       result,
		}
		tmpl.Execute(w, response)

	} else {
		userId := session.Values["Id"].(int)

		data, _ := connection.Conn.Query(context.Background(), "SELECT tb_projects.id, tb_projects.name, tb_projects.start_date, tb_projects.end_date, tb_projects.description, tb_projects.technologies,tb_projects.image FROM tb_projects LEFT JOIN tb_user ON tb_projects.user_id = tb_user.id WHERE tb_user.id=$1", userId)

		var result []Blog

		for data.Next() {

			var each = Blog{}
			err := data.Scan(&each.Id, &each.NameProject, &each.StarDate, &each.EndDate, &each.Message, &each.Tech, &each.Image)

			if err != nil {
				fmt.Println("Error baris 201 " + err.Error())
				return
			}

			hs := each.EndDate.Sub(each.StarDate).Hours()
			day, _ := math.Modf(hs / 24)
			bulan := int64(day / 30)
			tahun := int64(day / 365)

			if tahun > 0 {
				each.Duration = strconv.FormatInt(tahun, 10) + " Year"
			} else if bulan > 0 {
				each.Duration = strconv.FormatInt(bulan, 10) + " Month"
			} else {
				each.Duration = fmt.Sprintf("%.0f", day) + " Day"
			}
			result = append(result, each)
		}

		fmt.Println(result)
		response := map[string]interface{}{
			"DataSession": Data,
			"Blogs":       result,
		}

		tmpl.Execute(w, response)
	}
}

func detailProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var tmpl, err = template.ParseFiles("view/blog-detail.html")

	if err != nil {
		w.Write([]byte("messege: " + err.Error()))
		return
	}

	var ProjectDetail = Blog{}

	indexData, _ := strconv.Atoi(mux.Vars(r)["id"])

	err = connection.Conn.QueryRow(context.Background(), "SELECT id, name, start_date, end_date, description, technologies FROM tb_projects WHERE id=$1", indexData).Scan(&ProjectDetail.Id, &ProjectDetail.NameProject, &ProjectDetail.StarDate, &ProjectDetail.EndDate, &ProjectDetail.Message, &ProjectDetail.Tech, ProjectDetail.Image)

	if err != nil {
		fmt.Println("Error baris 168 " + err.Error())
		return
	}
	hs := ProjectDetail.EndDate.Sub(ProjectDetail.StarDate).Hours()
	day, _ := math.Modf(hs / 24)
	bulan := int64(day / 30)
	tahun := int64(day / 365)

	if tahun > 0 {
		ProjectDetail.Duration = strconv.FormatInt(tahun, 10) + " Year"
	} else if bulan > 0 {
		ProjectDetail.Duration = strconv.FormatInt(bulan, 10) + " Month"
	} else {
		ProjectDetail.Duration = fmt.Sprintf("%.0f", day) + " Day"
	}

	ProjectDetail.Format_starDate = ProjectDetail.StarDate.Format("2 January 2006")
	ProjectDetail.Format_endDate = ProjectDetail.EndDate.Format("2 January 2006")

	response := map[string]interface{}{
		"DataSession": Data,
		"Blogs":       ProjectDetail,
	}

	tmpl.Execute(w, response)
}

func formProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl, err := template.ParseFiles("view/add.html")

	if err != nil {
		w.Write([]byte("messege: " + err.Error()))
		return
	}

	response := map[string]interface{}{
		"DataSession": Data,
	}

	tmpl.Execute(w, response)
}

func contact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl, err := template.ParseFiles("view/contact.html")

	if err != nil {
		w.Write([]byte("messege: " + err.Error()))
		return
	}
	response := map[string]interface{}{
		"DataSession": Data,
	}

	tmpl.Execute(w, response)
}

func deleteProject(w http.ResponseWriter, r *http.Request) {
	indexDelete, _ := strconv.Atoi(mux.Vars(r)["id"])

	_, err := connection.Conn.Exec(context.Background(), "DELETE FROM tb_projects WHERE id=$1", indexDelete)

	if err != nil {
		w.Write([]byte("eror Baris 201 " + err.Error()))
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func editProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	tmpl, err := template.ParseFiles("view/edit-project.html")

	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	var EditProject = Blog{}
	indexData, _ := strconv.Atoi(mux.Vars(r)["id"])

	err = connection.Conn.QueryRow(context.Background(), "SELECT id, name, start_date, end_date, description, technologies,image FROM tb_projects WHERE id=$1", indexData).Scan(&EditProject.Id, &EditProject.NameProject, &EditProject.StarDate, &EditProject.EndDate, &EditProject.Message, &EditProject.Tech, &EditProject.Image)

	if err != nil {
		w.Write([]byte("error baris 223 " + err.Error()))
	}

	EditProject.Edit_starDate = EditProject.StarDate.Format("2006-01-02")
	EditProject.Edit_endDate = EditProject.EndDate.Format("2006-01-02")

	dataEdit := map[string]interface{}{
		"Edit": EditProject,
	}

	tmpl.Execute(w, dataEdit)
}

func updateProject(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()

	if err != nil {
		log.Fatal(err)
	}

	indexData, _ := strconv.Atoi(mux.Vars(r)["id"])

	// var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	// session, _ := store.Get(r, "SESSION_KEY")

	// userId := session.Values["Id"].(int)

	var nameProject = r.PostForm.Get("projectName")
	var startDate = r.PostForm.Get("startDate")
	var endDate = r.PostForm.Get("endDate")
	var desc = r.PostForm.Get("Description")

	dataContext := r.Context().Value("dataFile")
	image := dataContext.(string)

	var tech []string
	for key, values := range r.Form {
		for _, value := range values {
			if key == "technologies" {
				tech = append(tech, value)
			}
		}
	}

	_, err = connection.Conn.Exec(context.Background(), "UPDATE tb_projects SET name=$2,start_date=$3,end_date=$4,description=$5,technologies=$6, image=$7 WHERE id=$1", indexData, nameProject, startDate, endDate, desc, tech, image)

	if err != nil {
		w.Write([]byte("error Baris 267: " + err.Error()))
	}
	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl, err := template.ParseFiles("view/register.html")

	if err != nil {
		w.Write([]byte("messege: " + err.Error()))
		return
	}

	tmpl.Execute(w, nil)
}

func formRegister(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	var name = r.PostForm.Get("name")
	var email = r.PostForm.Get("email")
	var password = r.PostForm.Get("password")

	hasPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 10)

	_, err = connection.Conn.Exec(context.Background(), "INSERT INTO tb_user(name,email,password) VALUES ($1,$2,$3)", name, email, hasPassword)

	if err != nil {
		w.Write([]byte("error dibaris 302 " + err.Error()))
		return
	}

	http.Redirect(w, r, "/register", http.StatusMovedPermanently)
}

func login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl, err := template.ParseFiles("view/login.html")

	if err != nil {
		w.Write([]byte("messege: " + err.Error()))
		return
	}

	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")

	flasMessege := session.Flashes("dataSession")

	var flashes []string
	if len(flasMessege) > 0 {
		session.Save(r, w)
		for _, f1 := range flasMessege {
			flashes = append(flashes, f1.(string))
		}
	}

	Data.FlashData = strings.Join(flashes, "")
	tmpl.Execute(w, Data)
}

func formLogin(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	var email = r.PostForm.Get("email")
	var password = r.PostForm.Get("password")

	var user = User{}

	err = connection.Conn.QueryRow(context.Background(), "SELECT * FROM tb_user WHERE email=$1", email).Scan(&user.Id, &user.Name, &user.Email, &user.Password)

	if err != nil {
		var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
		session, _ := store.Get(r, "SESSION_KEY")

		session.AddFlash("Email salah atau Belum terdaftar!", "dataSession")
		session.Save(r, w)

		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))

	if err != nil {
		var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
		session, _ := store.Get(r, "SESSION_KEY")

		session.AddFlash("Password Salah!", "dataSession")
		session.Save(r, w)

		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
	}

	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")

	session.Values["Id"] = user.Id
	session.Values["Nama"] = user.Name
	session.Values["Email"] = user.Email
	session.Values["Login"] = true
	session.Options.MaxAge = 10800

	session.AddFlash("Login berhasil", "dataSession")
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
	// tmpl.Execute(w, nil)
}

func logout(w http.ResponseWriter, r *http.Request) {
	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")
	session.Options.MaxAge = -1
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
