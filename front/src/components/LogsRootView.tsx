import { ExploitState } from "@/proto/exploits/api";
import { LogLine } from "@/proto/logs/api";
import { useRootContext } from "@/routes/rootContext";
import { useExploitServiceClient } from "@/services/exploits";
import { useLogsServiceClient } from "@/services/logs";
import { useWindowDimensions } from "@/utils/window";
import {
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
  const [linesToken, setLinesToken] = useState<string>("");
  const [loading, setLoading] = useState<boolean>(false);
  const [hasMoreLines, setHasMoreLines] = useState<boolean>(true);
  const [linesRequested, setLinesRequested] = useState<boolean>(false);

  const prevStateRef = useRef<S>(state);

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
    const prevState = prevStateRef.current;
    if (
      state.exploitID !== prevState.exploitID ||
      state.exploitVersion !== prevState.exploitVersion
    ) {
      unstable_batchedUpdates(() => {
        setLines([]);
        setLinesToken("");
        setLoading(false);
        setHasMoreLines(true);
      });
      prevStateRef.current = state;
    }

    if (
      !state.exploitID ||
      !state.exploitVersion ||
      loading ||
      !linesRequested
    ) {
      return;
    }

    setLoading(true);
    const wrapper = async () => {
      const stream = logsServiceClient.searchLogLines({
        exploit: state.exploitID,
        version: state.exploitVersion,
        lastToken: linesToken,
      });

      const lines: LogLine[] = [];
      let nextToken = "";
      for await (const response of stream) {
        lines.push(...response.lines);
        nextToken = response.lastToken;
      }

      unstable_batchedUpdates(() => {
        setLines((prevLines) => [...prevLines, ...lines]);
        setLinesToken(nextToken);
        setLoading(false);
        setHasMoreLines(lines.length > 0);
        setLinesRequested(false);
      });
    };
    void wrapper();
  }, [state, logsServiceClient, loading, linesToken, linesRequested]);

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
        <LogsView
          lines={lines}
          viewHeight={viewSize}
          hasMoreLines={hasMoreLines}
          isLoading={loading}
          loadNext={() => {
            setLinesRequested(true);
          }}
        />
      </Grid>
    </Grid>
  );
}
