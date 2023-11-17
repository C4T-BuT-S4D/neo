import { createContext, useContext } from "react";

type RootContextType = { topbarRef: React.RefObject<HTMLElement> };

export const RootContext = createContext<RootContextType | null>(null);

export function useRootContext(): RootContextType {
  const ctx = useContext(RootContext);
  if (!ctx) {
    throw new Error("useRootContext must be used within RootContext.Provider");
  }
  return ctx;
}
