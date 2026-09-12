import { useEffect, useState } from "react";

// Returns a debounced copy of value that only updates after delay ms
// without changes. Used for search inputs to avoid a request per keystroke.
export function useDebouncedValue<T>(value: T, delay = 350): T {
  const [debounced, setDebounced] = useState(value);

  useEffect(() => {
    const timer = setTimeout(() => setDebounced(value), delay);
    return () => clearTimeout(timer);
  }, [value, delay]);

  return debounced;
}
