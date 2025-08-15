document.addEventListener('DOMContentLoaded', function() {
    const loginForm = document.getElementById('loginForm');
    
    if (loginForm) {
        loginForm.addEventListener('submit', async function(e) {
            e.preventDefault();
            
            const username = document.getElementById('username').value;
            const password = document.getElementById('password').value;
            
            try {
                const response = await fetch('/api/auth/login', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({ username, password })
                });
                
                if (response.ok) {
                    window.location.href = '/dashboard';
                } else {
                    const error = await response.json();
                    alert(error.message || 'Login failed');
                }
            } catch (err) {
                console.error('Login error:', err);
                alert('Login request failed');
            }
        });
    }
});

// Check authentication status
async function checkAuth() {
    try {
        const response = await fetch('/api/auth/check');
        if (!response.ok) {
            window.location.href = '/login';
        }
    } catch (err) {
        console.error('Auth check error:', err);
        window.location.href = '/login';
    }
}

// Initialize auth check for protected pages
if (window.location.pathname !== '/login') {
    checkAuth();
}