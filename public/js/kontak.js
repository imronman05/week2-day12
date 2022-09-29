function sendContact(){
    let name = document.getElementById('name').value;
    let email = document.getElementById('email').value;
    let phone = document.getElementById('phone').value;
    let subject = document.getElementById('subject').value;
    let message = document.getElementById('message').value;

    if(name == ' ' || email == ' ' || phone == ' ' || subject == '' || message == ''){
        return alert('Isikan form dengan benar!!!');
    }

    const emailReceiver = 'imronman1998@gmail.com';
    const a = document.createElement('a'); 

    a.href = `https://mail.google.com/mail/?view=cm&fs=1&to=${emailReceiver}&su=${subject}&body=hai perkenalkan nama saya ${name}, ${message}, berikut nomor yang bisa anda hubungi ${phone} atau email lain nya ${email}`;
    
    a.target = '_blank'; 
    a.click(); 

}

const buttonSend = document.getElementById('button-send');
    buttonSend.addEventListener('click',function(e){
        e.preventDefault() 
        sendContact()
    })