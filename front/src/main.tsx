import "@/index.css";
import ExploitsContainer from "@/routes/exploits";
import Root from "@/routes/root";
import theme from "@/theme";
import "@fontsource/roboto/300.css";
import "@fontsource/roboto/400.css";
import "@fontsource/roboto/500.css";
import "@fontsource/roboto/700.css";
import { ThemeProvider } from "@mui/material";
import React from "react";
import ReactDOM from "react-dom/client";
import { HelmetProvider } from "react-helmet-async";
import { createBrowserRouter, RouterProvider } from "react-router-dom";
import AuthProvider from "./routes/auth";
import LogsContainer from "./routes/logs";

const router = createBrowserRouter([
  {
    path: "/",
    element: <Root />,
    children: [
      {
        index: true,
        element: (
          <AuthProvider>
            <ExploitsContainer />
          </AuthProvider>
        ),
      },
      {
        path: "/logs",
        element: (
          <AuthProvider>
            <LogsContainer />
          </AuthProvider>
        ),
      },
    ],
  },
]);

// eslint-disable-next-line @typescript-eslint/no-non-null-assertion
ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <HelmetProvider>
      <ThemeProvider theme={theme}>
        <RouterProvider router={router} />
      </ThemeProvider>
    </HelmetProvider>
  </React.StrictMode>
);
