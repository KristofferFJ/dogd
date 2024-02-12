document.addEventListener('DOMContentLoaded', function () {
    const buttons = document.querySelectorAll('.question-btn');
    const modal = document.getElementById("answer");
    const span = document.getElementsByClassName("close")[0];
    const questionTextContainer = document.getElementById("questionText");
    const loader = document.getElementById("loader"); // Ensure you have a loader element in your HTML

    buttons.forEach(button => {
        button.addEventListener('click', function () {
            loader.style.display = "block"; // Show loader when the button is clicked
            modal.style.display = "block"; // Show the modal background immediately
            const endpoint = this.getAttribute('data-endpoint');

            fetch(`http://localhost:1337/${endpoint}`)
                .then(response => {
                    if (!response.ok) {
                        throw new Error('Network response was not ok');
                    }
                    return response.json(); // This reads the stream and converts it to JSON
                })
                .then(data => {
                    loader.style.display = "none"; // Hide loader when data is fetched
                    // Assuming 'Text' is your actual question text and 'Url' is the link you want to make clickable
                    questionTextContainer.innerHTML = `${data.Text} <br><br> <a href="${data.Url}" target="_blank">Svar</a>`;
                })
                .catch(error => {
                    console.error('There was a problem with your fetch operation:', error);
                    loader.style.display = "none"; // Ensure loader is hidden on error
                    alert('Failed to load question. Please try again later.');
                });
        });
    });

    // When the user clicks on <span> (x), close the modal
    span.onclick = function() {
        modal.style.display = "none";
    };

    // When the user clicks anywhere outside of the modal, close it
    window.onclick = function(event) {
        if (event.target == modal) {
            modal.style.display = "none";
        }
    };
});
