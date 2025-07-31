function validateForm() {
    const email = document.getElementById('email').value.trim();
    const username = document.getElementById('username').value.trim();
    const password = document.getElementById('password').value;

    // const emailRegex = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)+\.[a-zA-Z]{2,}$/;

    if (email.includes('..')) {
        alert('Email cannot contain consecutive dots.');
        return false;
    }
    if (email.startsWith('.') || email.endsWith('.')) {
        alert('Email cannot start or end with a dot.');
        return false;
    }
    if (email.includes('@.') || email.includes('.@')) {
        alert('Invalid email format.');
        return false;
    }

    const parts = email.split('@');
    if (parts.length !== 2) {
        alert('Invalid email format.');
        return false;
    }
    const domain = parts[1];
    // if (!domain.includes('.')) {
    //     alert('Email must have a valid domain (e.g., user@domain.com).');
    //     return false;
    // }
    const domainParts = domain.split('.');
    if (domainParts.length < 2 || domainParts[0] === '') {
        alert('Email must have a valid domain structure (e.g., user@domain.com).');
        return false;
    }

    if (email.length > 254) {
        alert('Please enter a valid email address.');
        return false;
    }

    if (username.length < 3) {
        alert('Username must be at least 3 characters long.');
        return false;
    }
    if (username.length > 20) {
        alert('Username must be no more than 20 characters long.');
        return false;
    }
    const usernameRegex = /^[a-zA-Z0-9_-]+$/;
    if (!usernameRegex.test(username)) {
        alert('Username can only contain letters, numbers, underscores, and hyphens.');
        return false;
    }

    if (password.length < 8) {
        alert('Password must be at least 8 characters long.');
        return false;
    }
    if (password.length > 50) {
        alert('Password must be no more than 50 characters long.');
        return false;
    }

    return true;
} 