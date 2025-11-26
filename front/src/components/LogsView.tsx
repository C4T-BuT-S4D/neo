import { LogLine } from "@/proto/logs/api";
import { CSSProperties } from "react";
import { ListOnItemsRenderedProps, VariableSizeList } from "react-window";
import InfiniteLoader from "react-window-infinite-loader";

interface P {
  lines: LogLine[];
  viewHeight: number;
  hasMoreLines: boolean;
  isLoading: boolean;
  loadNext: () => void;
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

  const isLineLoaded = (index: number) =>
    !props.hasMoreLines || index < props.lines.length;

  const lineCount = props.hasMoreLines
    ? props.lines.length + 1
    : props.lines.length;

  const loadMoreLines = props.isLoading ? () => {} : props.loadNext;

  const LineRow = ({
    index,
    style,
  }: {
    index: number;
    style: CSSProperties;
  }) => {
    if (!isLineLoaded(index)) {
      return (
        <div
          key={index}
          style={{
            ...style,
            color: "grey",
            fontSize: 12,
            lineHeight: "16px",
          }}
        >
          Loading...
        </div>
      );
    }

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

  const lineHeight = (index: number): number => {
    if (!isLineLoaded(index)) {
      return 16;
    }
    return lineCounts[index] * 16;
  };

  return (
    <InfiniteLoader
      isItemLoaded={isLineLoaded}
      itemCount={lineCount}
      loadMoreItems={loadMoreLines}
    >
      {({
        onItemsRendered,
        ref,
      }: {
        onItemsRendered: (props: ListOnItemsRenderedProps) => unknown;
        ref: (ref: unknown) => void;
      }) => (
        <VariableSizeList
          height={props.viewHeight}
          itemCount={lineCount}
          overscanCount={10}
          onItemsRendered={onItemsRendered}
          itemSize={lineHeight}
          ref={ref}
          width="100%"
        >
          {LineRow}
        </VariableSizeList>
      )}
    </InfiniteLoader>
  );
}
