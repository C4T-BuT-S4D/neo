import { AppBar, Box, Button, Toolbar } from "@mui/material";
import { useRef } from "react";
import { Link, Outlet } from "react-router-dom";
import { RootContext } from "./rootContext";

export default function Root() {
  const topbarRef = useRef<HTMLElement>(null);

  return (
    <>
      <AppBar position="static" ref={topbarRef}>
        <Toolbar>
          <Box
            component="img"
            sx={{
              height: 64,
            }}
            alt="Neo"
            src="/logo_toolbar.png"
          />
          <Button component={Link} to={"/"} color="inherit">
            Exploits
          </Button>
          <Button component={Link} to={"/logs"} color="inherit">
            Logs
          </Button>
        </Toolbar>
      </AppBar>
      <Box sx={{ width: "100%", height: "100%" }}>
        <RootContext.Provider value={{ topbarRef }}>
          <Outlet />
        </RootContext.Provider>
      </Box>
    </>
  );
}
