import LogsRootView from "@/components/LogsRootView";
import { ExploitState } from "@/proto/exploits/api";
import { useExploitServiceClient } from "@/services/exploits";
import { useEffect, useState } from "react";
import { Helmet } from "react-helmet-async";

interface S {
  exploits: ExploitState[];
}

export default function LogsContainer() {
  const exploitServiceClient = useExploitServiceClient();
  const [state, setState] = useState<S>();

  useEffect(() => {
    const fetchExploits = async () => {
      const response = await exploitServiceClient.ping({
        payload: { $case: "serverInfoRequest", serverInfoRequest: {} },
      });
      setState({
        exploits:
          response.state?.exploits.sort(
            (e1: ExploitState, e2: ExploitState) => {
              return e1.exploitId.localeCompare(e2.exploitId);
            }
          ) || [],
      });
    };
    void fetchExploits();
  }, [exploitServiceClient]);

  return (
    <>
      <Helmet>
        <title>Neo Logs</title>
      </Helmet>
      {state && <LogsRootView exploits={state.exploits} />}
    </>
  );
}
