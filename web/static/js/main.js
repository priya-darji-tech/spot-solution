// ==========================================================================
// NEXORA DYNAMICS - CLIENT-SIDE LOGIC & FORM WORKFLOWS
// ==========================================================================

document.addEventListener('DOMContentLoaded', () => {
    // 1. Mobile Menu Toggle
    const navToggle = document.getElementById('navToggle');
    const navMenu = document.getElementById('navMenu');

    if (navToggle && navMenu) {
        navToggle.addEventListener('click', () => {
            navMenu.classList.toggle('open');
            navToggle.classList.toggle('active');
        });
    }

    // 2. Helper to fetch cookie values (for CSRF token)
    function getCookie(name) {
        const value = `; ${document.cookie}`;
        const parts = value.split(`; ${name}=`);
        if (parts.length === 2) return parts.pop().split(';').shift();
        return '';
    }

    // 3. Contact Form Submission (Handles both Hero quick inquiry & dedicated Contact page form)
    const contactForms = document.querySelectorAll('.ajax-contact-form');

    contactForms.forEach(form => {
        form.addEventListener('submit', async (e) => {
            e.preventDefault();

            const submitBtn = form.querySelector('button[type="submit"]');
            const alertSuccess = form.querySelector('.alert-success-msg');
            const alertError = form.querySelector('.alert-error-msg');
            const originalBtnText = submitBtn.innerHTML;

            // Reset alerts
            if (alertSuccess) alertSuccess.style.display = 'none';
            if (alertError) alertError.style.display = 'none';

            // Extract form data
            const formData = new FormData(form);
            const payload = Object.fromEntries(formData.entries());

            // Handle name splitting if last_name is not explicitly provided
            if (!payload.last_name && payload.first_name) {
                const nameParts = payload.first_name.trim().split(/\s+/);
                if (nameParts.length > 1) {
                    payload.first_name = nameParts[0];
                    payload.last_name = nameParts.slice(1).join(' ');
                } else {
                    payload.last_name = 'Client';
                }
            }

            // Ensure CSRF token is present
            const csrfToken = payload.csrf_token || getCookie('csrf_token');

            // Button loading state
            submitBtn.disabled = true;
            submitBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Processing Inquiry...';

            try {
                const response = await fetch('/api/contact', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'X-CSRF-Token': csrfToken
                    },
                    body: JSON.stringify(payload)
                });

                const data = await response.json();

                if (response.ok && data.success) {
                    if (alertSuccess) {
                        alertSuccess.textContent = data.message || 'Inquiry submitted successfully! We will follow up shortly.';
                        alertSuccess.style.display = 'block';
                    }
                    form.reset();
                } else {
                    if (alertError) {
                        alertError.textContent = data.error || 'Failed to submit inquiry. Please verify the form inputs.';
                        alertError.style.display = 'block';
                    }
                }
            } catch (err) {
                if (alertError) {
                    alertError.textContent = 'A network error occurred. Please try again or call us directly.';
                    alertError.style.display = 'block';
                }
            } finally {
                submitBtn.disabled = false;
                submitBtn.innerHTML = originalBtnText;
            }
        });
    });

    // 4. Admin Login Handler
    const loginForm = document.getElementById('adminLoginForm');
    if (loginForm) {
        loginForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            const emailInput = document.getElementById('loginEmail');
            const passInput = document.getElementById('loginPassword');
            const errorBox = document.getElementById('loginErrorBox');
            const submitBtn = loginForm.querySelector('button[type="submit"]');

            if (errorBox) errorBox.style.display = 'none';
            submitBtn.disabled = true;
            submitBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Authenticating...';

            try {
                const res = await fetch('/api/admin/login', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        email: emailInput.value.trim(),
                        password: passInput.value
                    })
                });

                const data = await res.json();
                if (res.ok && data.success) {
                    window.location.href = data.redirect || '/admin/leads';
                } else {
                    if (errorBox) {
                        errorBox.textContent = data.error || 'Invalid credentials';
                        errorBox.style.display = 'block';
                    }
                    submitBtn.disabled = false;
                    submitBtn.innerHTML = 'Sign In to Dashboard';
                }
            } catch (err) {
                if (errorBox) {
                    errorBox.textContent = 'Server connection failed';
                    errorBox.style.display = 'block';
                }
                submitBtn.disabled = false;
                submitBtn.innerHTML = 'Sign In to Dashboard';
            }
        });
    }

    // 5. Admin Logout Handler
    const logoutBtn = document.getElementById('adminLogoutBtn');
    if (logoutBtn) {
        logoutBtn.addEventListener('click', async (e) => {
            e.preventDefault();
            const csrfToken = getCookie('csrf_token');
            try {
                await fetch('/api/admin/logout', {
                    method: 'POST',
                    headers: { 'X-CSRF-Token': csrfToken }
                });
            } finally {
                window.location.href = '/admin/login';
            }
        });
    }
});
