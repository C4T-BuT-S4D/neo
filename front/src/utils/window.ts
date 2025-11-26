import { useSyncExternalStore } from "react";

export function useWindowDimensions() {
  return useSyncExternalStore(subscribe, getSnapshot);
}

function subscribe(callback: {
  (this: Window, ev: UIEvent): void;
  (this: Window, ev: UIEvent): void;
}) {
  window.addEventListener("resize", callback);
  return () => {
    window.removeEventListener("resize", callback);
  };
}

interface Snapshot {
  width: number;
  height: number;
}

let lastSnapshot: Snapshot | null = null;

function getSnapshot() {
  const newSnapshot = { width: window.innerWidth, height: window.innerHeight };
  if (
    lastSnapshot == null ||
    newSnapshot.width != lastSnapshot.width ||
    newSnapshot.height != lastSnapshot.height
  ) {
    lastSnapshot = newSnapshot;
  }
  return lastSnapshot;
}
