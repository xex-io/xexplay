import axios from "axios";

// Play API client — for all Play-specific endpoints (events, cards, users, etc.)
const apiClient = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/v1",
  headers: {
    "Content-Type": "application/json",
  },
});

apiClient.interceptors.request.use((config) => {
  if (typeof window !== "undefined") {
    const token = localStorage.getItem("admin_token");
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
  }
  return config;
});

apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (
      error.response?.status === 401 &&
      typeof window !== "undefined" &&
      !window.location.pathname.includes("/login")
    ) {
      localStorage.removeItem("admin_token");
      localStorage.removeItem("admin_user");
      window.location.href = "/login";
    }
    return Promise.reject(error);
  }
);

// Exchange API client — for admin authentication
const exchangeApiUrl =
  process.env.NEXT_PUBLIC_EXCHANGE_API_URL || "https://api.xex.to";

export const exchangeClient = axios.create({
  baseURL: exchangeApiUrl,
  headers: {
    "Content-Type": "application/json",
  },
});

export default apiClient;
