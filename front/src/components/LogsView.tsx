import { LogLine } from "@/proto/logs/api";
import { CSSProperties } from "react";
import { VariableSizeList } from "react-window";

interface P {
  lines: LogLine[];
  viewHeight: number;
}

export default function LogsView(props: P) {
  const mapLevelColor = (level: string) => {
    switch (level) {
      case "debug":
        return "grey";
      case "info":
        return "black";
      case "warning":
        return "orange";
      case "error":
        return "red";
      default:
        return "black";
    }
  };

  const lineCounts = props.lines.map((line) => line.message.split("\n").length);

  const LineRow = ({
    index,
    style,
  }: {
    index: number;
    style: CSSProperties;
  }) => {
    const line = props.lines[index];

    return (
      <pre
        key={index}
        style={{
          ...style,
          color: mapLevelColor(line.level),
          fontSize: 12,
          lineHeight: "16px",
        }}
      >
        [{line.team}] [{line.level}] {line.message}
      </pre>
    );
  };

  return (
    <VariableSizeList
      height={props.viewHeight}
      itemCount={props.lines.length}
      overscanCount={10}
      itemSize={(index) => lineCounts[index] * 16}
      width="100%"
    >
      {LineRow}
    </VariableSizeList>
  );
}
