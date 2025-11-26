import { Duration } from "@/proto/google/protobuf/duration";
import Long from "long";

export function formatDuration(d: Duration): string {
  return formatSeconds(d.seconds.toInt());
}

export function formatSeconds(s: number): string {
  const hours = Math.floor(s / 3600);
  const minutes = Math.floor((s % 3600) / 60);
  const seconds = Math.floor(s % 60);

  let res = "";
  if (hours > 0) {
    res += `${hours}h`;
  }
  if (minutes > 0) {
    res += `${minutes}m`;
  }
  if (seconds > 0) {
    res += `${seconds}s`;
  }
  if (res.length === 0) {
    res = "0s";
  }
  return res;
}

export function parseDuration(s: string): Duration {
  const numRegex = /^\d+$/;
  if (numRegex.test(s)) {
    return Duration.create({
      seconds: Long.fromNumber(parseInt(s)),
      nanos: 0,
    });
  }

  const regex =
    /^(?<hours>(\d+)h)?\s*(?<minutes>(\d+)m)?\s*(?<seconds>(\d+)s)?$/;
  const match = regex.exec(s);
  if (match === null) {
    throw new Error(`invalid duration: ${s}`);
  }

  interface Parsed {
    hours: string;
    minutes: string;
    seconds: string;
  }
  const parsed: Partial<Parsed> | undefined = match.groups;

  const hours = parsed?.hours === undefined ? 0 : parseInt(parsed.hours);
  const minutes = parsed?.minutes === undefined ? 0 : parseInt(parsed.minutes);
  const seconds = parsed?.seconds === undefined ? 0 : parseInt(parsed.seconds);

  return Duration.create({
    seconds: Long.fromNumber(hours * 3600 + minutes * 60 + seconds),
    nanos: 0,
  });
}
