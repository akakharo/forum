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