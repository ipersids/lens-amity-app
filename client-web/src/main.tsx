import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter } from "react-router";
import App from "./App.tsx";
import SessionExpiryHandler from "./features/auth/components/SessionExpiryHandler.tsx";

const queryClient = new QueryClient();

createRoot(document.getElementById("root") as HTMLElement).render(
  <StrictMode>
    <BrowserRouter>
      <QueryClientProvider client={queryClient}>
        <SessionExpiryHandler />
        <App />
      </QueryClientProvider>
    </BrowserRouter>
  </StrictMode>,
);
