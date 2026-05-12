export const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";

async function request(path: string, options: RequestInit = {}) {
  const token = typeof window !== "undefined" ? localStorage.getItem("token") : null;

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers as Record<string, string> || {}),
  };

  if (token) {
    headers["Authorization"] = token;
  }

  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers,
  });

  if (!res.ok) {
    const error = await res.json().catch(() => ({ error: "Request failed" }));
    throw new Error(error.error || `HTTP ${res.status}`);
  }

  return res.json();
}

export const api = {
  auth: {
    register: (data: { email: string; password: string; first_name: string; last_name: string; phone?: string }) =>
      request("/identity/register", { method: "POST", body: JSON.stringify(data) }),
    login: (email: string, password: string) =>
      request("/identity/login", { method: "POST", body: JSON.stringify({ email, password }) }),
  },
  profile: {
    get: (id: string) => request(`/identity/profile/${id}`),
    update: (id: string, data: Record<string, string>) =>
      request(`/identity/profile/${id}`, { method: "PUT", body: JSON.stringify(data) }),
  },
  vehicles: {
    list: (page = 1, limit = 10) => request(`/vehicles?page=${page}&limit=${limit}`),
    get: (id: string) => request(`/vehicles/${id}`),
  },
  bookings: {
    list: (userId: string) => request(`/users/${userId}/bookings`),
    create: (data: Record<string, string>) => request("/bookings", { method: "POST", body: JSON.stringify(data) }),
    get: (id: string) => request(`/bookings/${id}`),
    cancel: (id: string, reason: string) => request(`/bookings/${id}`, { method: "DELETE", body: JSON.stringify({ reason }) }),
    availability: (vehicleId: string, startDate: string, endDate: string) =>
      request(`/bookings/availability?vehicle_id=${vehicleId}&start_date=${startDate}&end_date=${endDate}`),
    calculatePrice: (vehicleId: string, startDate: string, endDate: string) =>
      request("/bookings/calculate-price", { method: "POST", body: JSON.stringify({ vehicle_id: vehicleId, start_date: startDate, end_date: endDate }) }),
  },
};
