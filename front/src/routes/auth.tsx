import { rawExploitServiceClient } from "@/services/exploits";
import { usePersistentStorageValue } from "@/utils/storage";
import { Box, TextField } from "@mui/material";
import { ClientError, Metadata } from "nice-grpc-web";
import { useEffect, useState } from "react";
import ReactDOM from "react-dom";
import { Helmet } from "react-helmet-async";
import { AuthContext } from "./authContext";

export default function AuthProvider({
  children,
}: {
  children: React.ReactNode;
}) {
  const [error, setError] = useState<string | null>(null);
  const [token, setToken] = usePersistentStorageValue<string | null>("token");
  const [checkedToken, setCheckedToken] = useState<boolean>(false);

  useEffect(() => {
    const checkToken = async (token: string) => {
      try {
        await rawExploitServiceClient.ping(
          {
            payload: { $case: "serverInfoRequest", serverInfoRequest: {} },
          },
          { metadata: new Metadata({ authorization: token }) }
        );
        ReactDOM.unstable_batchedUpdates(() => {
          setToken(token);
          setCheckedToken(true);
          setError(null);
        });
      } catch (e) {
        let message = "Unknown error";
        if (e instanceof ClientError) {
          message = e.message;
        } else if (e instanceof Error) {
          message = e.message;
        }

        ReactDOM.unstable_batchedUpdates(() => {
          setCheckedToken(false);
          setError(message);
        });
      }
    };

    if (token) {
      void checkToken(token);
    }
  }, [token, setToken]);

  const handleTokenChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setToken(event.target.value);
  };

  if (!token || !checkedToken) {
    return (
      <>
        <Helmet>
          <title>Neo</title>
        </Helmet>
        <Box padding={2}>
          <TextField
            label="Token"
            variant="outlined"
            size="small"
            fullWidth
            error={!!error}
            helperText={error}
            onChange={handleTokenChange}
            value={token ?? ""}
          />
        </Box>
      </>
    );
  }

  return (
    <AuthContext.Provider
      value={{ metadata: new Metadata({ authorization: token }) }}
    >
      {children}
    </AuthContext.Provider>
  );
}
