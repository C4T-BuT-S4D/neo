import { ExploitState } from "@/proto/exploits/api";
import { LogLine } from "@/proto/logs/api";
import { useRootContext } from "@/routes/rootContext";
import { useExploitServiceClient } from "@/services/exploits";
import { useLogsServiceClient } from "@/services/logs";
import { useWindowDimensions } from "@/utils/window";
import {
  CircularProgress,
  FormControl,
  Grid,
  InputLabel,
  MenuItem,
  Select,
  SelectChangeEvent,
  Stack,
  TextField,
} from "@mui/material";
import { useEffect, useRef, useState } from "react";
import { unstable_batchedUpdates } from "react-dom";
import LogsView from "./LogsView";

interface S {
  exploitID: string;
  exploitVersion: string;
}

interface P {
  exploits: ExploitState[];
}

export default function LogsRootView(props: P) {
  const [state, setState] = useState<S>({ exploitID: "", exploitVersion: "" });
  const [lines, setLines] = useState<LogLine[]>([]);
  const [loading, setLoading] = useState<boolean>(false);

  const exploitServiceClient = useExploitServiceClient();
  const logsServiceClient = useLogsServiceClient();

  const fetchExploitData = async (exploitID: string) => {
    try {
      const response = await exploitServiceClient.exploit({
        exploitId: exploitID,
      });
      return response.state?.version.toString() || "";
    } catch {
      return "";
    }
  };

  const handleExploitIDChange = (event: SelectChangeEvent) => {
    const wrapper = async () => {
      const exploitVersion = await fetchExploitData(event.target.value);
      setState({ exploitID: event.target.value, exploitVersion });
    };
    void wrapper();
  };

  const handleExploitVersionChange = (
    event: React.ChangeEvent<HTMLInputElement>
  ) => {
    setState({ ...state, exploitVersion: event.target.value });
  };

  useEffect(() => {
    const tailLogs = async () => {
      if (state.exploitID === "" || state.exploitVersion === "") {
        return;
      }
      setLoading(true);
      const stream = logsServiceClient.searchLogLines({
        exploit: state.exploitID,
        version: state.exploitVersion,
      });
      const lines: LogLine[] = [];
      for await (const response of stream) {
        lines.push(...response.lines);
      }
      unstable_batchedUpdates(() => {
        setLoading(false);
        setLines(lines);
      });
    };

    void tailLogs();
  }, [state.exploitID, state.exploitVersion, logsServiceClient]);

  const streamFormRef = useRef<HTMLDivElement>(null);
  const { topbarRef } = useRootContext();

  const { height: windowHeight } = useWindowDimensions();
  const viewSize =
    windowHeight -
    (topbarRef.current?.clientHeight ?? 0) -
    (streamFormRef.current?.clientHeight ?? 0) -
    1;

  return (
    <Grid spacing={2}>
      <Grid item ref={streamFormRef}>
        <Stack spacing={2} padding={2} direction="row">
          <FormControl fullWidth>
            <InputLabel id="demo-simple-select-label">Exploit</InputLabel>
            <Select
              labelId="demo-simple-select-label"
              id="demo-simple-select"
              label="Exploit"
              onChange={handleExploitIDChange}
            >
              {props.exploits.map((exploit) => (
                <MenuItem key={exploit.exploitId} value={exploit.exploitId}>
                  {exploit.exploitId}
                </MenuItem>
              ))}
            </Select>
          </FormControl>
          <TextField
            label="Version"
            variant="outlined"
            type="number"
            value={state.exploitVersion}
            fullWidth
            onChange={handleExploitVersionChange}
          />
        </Stack>
      </Grid>
      <Grid item paddingX={2}>
        {loading && <CircularProgress />}
        {!loading && <LogsView lines={lines} viewHeight={viewSize} />}
      </Grid>
    </Grid>
  );
}
