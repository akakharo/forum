document.querySelector('form').addEventListener('submit', function(e) {
    const checkboxes = document.querySelectorAll('input[name="category_id"]:checked');
    if (checkboxes.length === 0) {
        e.preventDefault();
        alert('Please select at least one category');
        return false;
    }
}); 