// API utility functions for handling requests with 401 automatic logout

export function handleUnauthorized(): void {
  // Clear all auth tokens
  localStorage.removeItem('token');
  localStorage.removeItem('user');
  sessionStorage.removeItem('token');
  sessionStorage.removeItem('user');

  // Redirect to login page
  window.location.href = '/login';
}

export async function apiFetch(input: RequestInfo | URL, init?: RequestInit): Promise<Response> {
  const response = await fetch(input, init);

  // Check for 401 Unauthorized
  if (response.status === 401) {
    console.warn('[API] Received 401 Unauthorized - clearing token and redirecting to login');
    handleUnauthorized();
    // Throw error to prevent further processing
    throw new Error('Unauthorized - redirecting to login');
  }

  return response;
}

export function getAuthHeaders(): HeadersInit {
  const token = localStorage.getItem('token') || sessionStorage.getItem('token');
  return {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  };
}
