// Define global variables
const apiUrl = 'http://localhost:8080/api';
const fileInput = document.querySelector('#file-input');
const fileList = document.querySelector('#file-list');

// Add event listeners
fileInput.addEventListener('change', handleFileUpload);

// Define event handlers
function handleFileUpload(event) {
  const file = event.target.files[0];
  const formData = new FormData();
  formData.append('file', file);
  fetch(apiUrl + '/upload', {
    method: 'POST',
    body: formData
  })
  .then(response => response.json())
  .then(data => {
    console.log(data);
    // Add the uploaded file to the file list
    const li = document.createElement('li');
    const a = document.createElement('a');
    a.href = apiUrl + '/file/' + data.filename;
    a.textContent = data.filename;
    li.appendChild(a);
    fileList.appendChild(li);
  })
  .catch(error => console.error(error));
}