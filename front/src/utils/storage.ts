import { useEffect, useState } from "react";

export function usePersistentStorageValue<T>(key: string, initialValue?: T) {
  const [value, setValue] = useState<T | undefined>(() => {
    const valueFromStorage = window.localStorage.getItem(key);
    if (valueFromStorage) {
      return JSON.parse(valueFromStorage) as unknown as T;
    }
    return initialValue;
  });

  useEffect(() => {
    if (value) {
      window.localStorage.setItem(key, JSON.stringify(value));
    }
  }, [key, value]);

  return [value, setValue] as const;
}
