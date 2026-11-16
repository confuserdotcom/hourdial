// Motion that code has to drive: a number counting up to its value. The
// rest is CSS (styles.css, .anim-*), and all of it rests under
// prefers-reduced-motion.
import { useEffect, useRef, useState } from "react";

/** easeOut is expo-out: fast off the mark, long settle. */
export function easeOut(t: number): number {
  return t >= 1 ? 1 : t <= 0 ? 0 : 1 - Math.pow(2, -10 * t);
}

/** tween is the value between from and to at t in 0–1, eased. */
export function tween(from: number, to: number, t: number): number {
  return from + (to - from) * easeOut(t);
}

export function reducedMotion(): boolean {
  return typeof matchMedia === "function" && matchMedia("(prefers-reduced-motion: reduce)").matches;
}

/** useCountUp runs from the last value shown to target in ms, so a number
 *  that changes rolls to its new value; at once under reduced motion. */
export function useCountUp(target: number, ms = 900): number {
  const [v, setV] = useState(() => (reducedMotion() ? target : 0));
  const shown = useRef(v);
  useEffect(() => {
    if (reducedMotion() || typeof requestAnimationFrame !== "function") {
      shown.current = target;
      setV(target);
      return;
    }
    const from = shown.current;
    const t0 = performance.now();
    let id = 0;
    const step = (now: number) => {
      const t = Math.min(1, (now - t0) / ms);
      shown.current = tween(from, target, t);
      setV(shown.current);
      if (t < 1) id = requestAnimationFrame(step);
    };
    id = requestAnimationFrame(step);
    return () => cancelAnimationFrame(id);
  }, [target, ms]);
  return v;
}
