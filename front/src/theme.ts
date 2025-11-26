import { createTheme } from "@mui/material/styles";

declare module "@mui/material/styles" {
  interface Theme {
    typography: {
      fontFamily: string;
    };
    palette: {
      primary: {
        main: string;
        contrastText: string;
      };
    };
  }
}

const theme = createTheme({
  typography: {
    fontFamily: "Monospace",
  },
  palette: {
    primary: {
      main: "#333333",
    },
  },
});

export default theme;
