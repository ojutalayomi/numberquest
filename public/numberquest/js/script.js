const socket = io({
  reconnectionDelayMax: 100000
});

socket.on('connect', () => {
  console.log('Connected'); // Log the connection status message
    socket.emit('register', 'ayomide'); // Emit 'register' event with user's ID
    console.log('Registered with ID:');
});

socket.on('notice', (reason) => {
  var mob = document.querySelector('.header p');
  mob.innerText = reason;
  mob.style.borderColor = "#15202b";
  console.log('Message:', reason);
});

socket.on('board', (reason) => {
  var mob = document.querySelector('h5');
  mob.innerText = reason;
  mob.style.borderColor = "#15202b";
  console.log('Message:', reason);
});

socket.on('prevMessages', (msg) => {
  if(!msg) return;
  console.log('prevMessages:', msg);
});

socket.on('reply', (msg) => {
  console.log('messages:', msg);
});

socket.on('disconnect', (reason) => {
  console.log('Disconnected:', reason);
});

socket.on('connect_error', (err) => {
  console.error('Connection Error:', err);
  if(err.message.includes("P: websocket error at tt.onError") && user?.id){
    socket.emit('register', 'ayomide'); // Emit 'register' event with user's ID
    console.log('Registering with ID:', 'ayomide');
  }
});

socket.on('error', (err) => {
  console.error('Socket Error:', err);
});

var pi = document.querySelector("input");
var span = document.querySelectorAll(".button");

async function sendInstruction(){
  var pin = pi.value;
  var mob = document.querySelector('h5');
  if (!pin) {
    mob.innerText = "Please enter a number";
    pi.style.borderColor = "#15202b";
    return false;
  }
  document.getElementById("value").value = '';
  const response = await fetch('/api?option='+pin)
  let data;
  try {
    const contentType = response.headers.get("content-type");
    if (contentType && contentType.includes("application/json")) {
      data = await response.json();
    } else {
      const text = await response.text();
      data = {
        message: text,
        success: false,
        gameOver: false
      };
    }
  } catch (e) {
    data = {
      message: "Error connecting to game",
      success: false, 
      gameOver: false
    };
  }
  // mob.innerText = data.message;
  if (data.gameOver) {
    pi.value = '';
  }
}

let txt = "Guess";
let msg = document.getElementById("msg");
let i = 0;

let typing = () =>{
  if(i < txt.length){
    setInterval(
      function(){
        msg.textContent += txt.charAt(i);
        i++
      },50)
  }
}
typing();