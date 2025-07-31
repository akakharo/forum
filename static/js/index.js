// Save scroll position before submitting like/dislike forms
const likeForms = document.querySelectorAll('form[action="/like"]');
likeForms.forEach(form => {
    form.addEventListener('submit', function() {
        sessionStorage.setItem('scrollPos', window.scrollY);
    });
});
// Restore scroll position after reload
window.addEventListener('load', function() {
    const scrollPos = sessionStorage.getItem('scrollPos');
    if (scrollPos) {
        window.scrollTo(0, parseInt(scrollPos));
        sessionStorage.removeItem('scrollPos');
    }
});

// Function to navigate to post page
function goToPost(postId) {
    window.location.href = '/post?id=' + postId;
}

// Prevent clicks on buttons and forms from triggering the post card click
document.addEventListener('click', function(e) {
    if (e.target.tagName === 'BUTTON' || e.target.tagName === 'FORM' || e.target.closest('form') || e.target.closest('button')) {
        e.stopPropagation();
    }
}); 