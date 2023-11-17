// @ts-nocheck
/* eslint-disable */
import Long from "long";
import type { CallContext, CallOptions } from "nice-grpc-common";
import _m0 from "protobufjs/minimal";
import { Empty } from "../google/protobuf/empty";

export const protobufPackage = "logs";

export interface LogLine {
  exploit: string;
  version: Long;
  message: string;
  level: string;
  team: string;
}

export interface AddLogLinesRequest {
  lines: LogLine[];
}

export interface SearchLogLinesRequest {
  exploit: string;
  version: Long;
}

export interface SearchLogLinesResponse {
  lines: LogLine[];
}

function createBaseLogLine(): LogLine {
  return { exploit: "", version: Long.ZERO, message: "", level: "", team: "" };
}

export const LogLine = {
  encode(message: LogLine, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.exploit !== "") {
      writer.uint32(10).string(message.exploit);
    }
    if (!message.version.isZero()) {
      writer.uint32(16).int64(message.version);
    }
    if (message.message !== "") {
      writer.uint32(26).string(message.message);
    }
    if (message.level !== "") {
      writer.uint32(34).string(message.level);
    }
    if (message.team !== "") {
      writer.uint32(42).string(message.team);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): LogLine {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseLogLine();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.exploit = reader.string();
          continue;
        case 2:
          if (tag !== 16) {
            break;
          }

          message.version = reader.int64() as Long;
          continue;
        case 3:
          if (tag !== 26) {
            break;
          }

          message.message = reader.string();
          continue;
        case 4:
          if (tag !== 34) {
            break;
          }

          message.level = reader.string();
          continue;
        case 5:
          if (tag !== 42) {
            break;
          }

          message.team = reader.string();
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): LogLine {
    return {
      exploit: isSet(object.exploit) ? globalThis.String(object.exploit) : "",
      version: isSet(object.version) ? Long.fromValue(object.version) : Long.ZERO,
      message: isSet(object.message) ? globalThis.String(object.message) : "",
      level: isSet(object.level) ? globalThis.String(object.level) : "",
      team: isSet(object.team) ? globalThis.String(object.team) : "",
    };
  },

  toJSON(message: LogLine): unknown {
    const obj: any = {};
    if (message.exploit !== "") {
      obj.exploit = message.exploit;
    }
    if (!message.version.isZero()) {
      obj.version = (message.version || Long.ZERO).toString();
    }
    if (message.message !== "") {
      obj.message = message.message;
    }
    if (message.level !== "") {
      obj.level = message.level;
    }
    if (message.team !== "") {
      obj.team = message.team;
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<LogLine>, I>>(base?: I): LogLine {
    return LogLine.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<LogLine>, I>>(object: I): LogLine {
    const message = createBaseLogLine();
    message.exploit = object.exploit ?? "";
    message.version = (object.version !== undefined && object.version !== null)
      ? Long.fromValue(object.version)
      : Long.ZERO;
    message.message = object.message ?? "";
    message.level = object.level ?? "";
    message.team = object.team ?? "";
    return message;
  },
};

function createBaseAddLogLinesRequest(): AddLogLinesRequest {
  return { lines: [] };
}

export const AddLogLinesRequest = {
  encode(message: AddLogLinesRequest, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    for (const v of message.lines) {
      LogLine.encode(v!, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): AddLogLinesRequest {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseAddLogLinesRequest();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.lines.push(LogLine.decode(reader, reader.uint32()));
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): AddLogLinesRequest {
    return { lines: globalThis.Array.isArray(object?.lines) ? object.lines.map((e: any) => LogLine.fromJSON(e)) : [] };
  },

  toJSON(message: AddLogLinesRequest): unknown {
    const obj: any = {};
    if (message.lines?.length) {
      obj.lines = message.lines.map((e) => LogLine.toJSON(e));
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<AddLogLinesRequest>, I>>(base?: I): AddLogLinesRequest {
    return AddLogLinesRequest.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<AddLogLinesRequest>, I>>(object: I): AddLogLinesRequest {
    const message = createBaseAddLogLinesRequest();
    message.lines = object.lines?.map((e) => LogLine.fromPartial(e)) || [];
    return message;
  },
};

function createBaseSearchLogLinesRequest(): SearchLogLinesRequest {
  return { exploit: "", version: Long.ZERO };
}

export const SearchLogLinesRequest = {
  encode(message: SearchLogLinesRequest, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.exploit !== "") {
      writer.uint32(10).string(message.exploit);
    }
    if (!message.version.isZero()) {
      writer.uint32(16).int64(message.version);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): SearchLogLinesRequest {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseSearchLogLinesRequest();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.exploit = reader.string();
          continue;
        case 2:
          if (tag !== 16) {
            break;
          }

          message.version = reader.int64() as Long;
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): SearchLogLinesRequest {
    return {
      exploit: isSet(object.exploit) ? globalThis.String(object.exploit) : "",
      version: isSet(object.version) ? Long.fromValue(object.version) : Long.ZERO,
    };
  },

  toJSON(message: SearchLogLinesRequest): unknown {
    const obj: any = {};
    if (message.exploit !== "") {
      obj.exploit = message.exploit;
    }
    if (!message.version.isZero()) {
      obj.version = (message.version || Long.ZERO).toString();
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<SearchLogLinesRequest>, I>>(base?: I): SearchLogLinesRequest {
    return SearchLogLinesRequest.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<SearchLogLinesRequest>, I>>(object: I): SearchLogLinesRequest {
    const message = createBaseSearchLogLinesRequest();
    message.exploit = object.exploit ?? "";
    message.version = (object.version !== undefined && object.version !== null)
      ? Long.fromValue(object.version)
      : Long.ZERO;
    return message;
  },
};

function createBaseSearchLogLinesResponse(): SearchLogLinesResponse {
  return { lines: [] };
}

export const SearchLogLinesResponse = {
  encode(message: SearchLogLinesResponse, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    for (const v of message.lines) {
      LogLine.encode(v!, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): SearchLogLinesResponse {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseSearchLogLinesResponse();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.lines.push(LogLine.decode(reader, reader.uint32()));
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): SearchLogLinesResponse {
    return { lines: globalThis.Array.isArray(object?.lines) ? object.lines.map((e: any) => LogLine.fromJSON(e)) : [] };
  },

  toJSON(message: SearchLogLinesResponse): unknown {
    const obj: any = {};
    if (message.lines?.length) {
      obj.lines = message.lines.map((e) => LogLine.toJSON(e));
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<SearchLogLinesResponse>, I>>(base?: I): SearchLogLinesResponse {
    return SearchLogLinesResponse.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<SearchLogLinesResponse>, I>>(object: I): SearchLogLinesResponse {
    const message = createBaseSearchLogLinesResponse();
    message.lines = object.lines?.map((e) => LogLine.fromPartial(e)) || [];
    return message;
  },
};

export type ServiceDefinition = typeof ServiceDefinition;
export const ServiceDefinition = {
  name: "Service",
  fullName: "logs.Service",
  methods: {
    addLogLines: {
      name: "AddLogLines",
      requestType: AddLogLinesRequest,
      requestStream: false,
      responseType: Empty,
      responseStream: false,
      options: {},
    },
    searchLogLines: {
      name: "SearchLogLines",
      requestType: SearchLogLinesRequest,
      requestStream: false,
      responseType: SearchLogLinesResponse,
      responseStream: true,
      options: {},
    },
  },
} as const;

export interface ServiceImplementation<CallContextExt = {}> {
  addLogLines(request: AddLogLinesRequest, context: CallContext & CallContextExt): Promise<DeepPartial<Empty>>;
  searchLogLines(
    request: SearchLogLinesRequest,
    context: CallContext & CallContextExt,
  ): ServerStreamingMethodResult<DeepPartial<SearchLogLinesResponse>>;
}

export interface ServiceClient<CallOptionsExt = {}> {
  addLogLines(request: DeepPartial<AddLogLinesRequest>, options?: CallOptions & CallOptionsExt): Promise<Empty>;
  searchLogLines(
    request: DeepPartial<SearchLogLinesRequest>,
    options?: CallOptions & CallOptionsExt,
  ): AsyncIterable<SearchLogLinesResponse>;
}

type Builtin = Date | Function | Uint8Array | string | number | boolean | undefined;

export type DeepPartial<T> = T extends Builtin ? T
  : T extends Long ? string | number | Long : T extends globalThis.Array<infer U> ? globalThis.Array<DeepPartial<U>>
  : T extends ReadonlyArray<infer U> ? ReadonlyArray<DeepPartial<U>>
  : T extends { $case: string } ? { [K in keyof Omit<T, "$case">]?: DeepPartial<T[K]> } & { $case: T["$case"] }
  : T extends {} ? { [K in keyof T]?: DeepPartial<T[K]> }
  : Partial<T>;

type KeysOfUnion<T> = T extends T ? keyof T : never;
export type Exact<P, I extends P> = P extends Builtin ? P
  : P & { [K in keyof P]: Exact<P[K], I[K]> } & { [K in Exclude<keyof I, KeysOfUnion<P>>]: never };

if (_m0.util.Long !== Long) {
  _m0.util.Long = Long as any;
  _m0.configure();
}

function isSet(value: any): boolean {
  return value !== null && value !== undefined;
}

export type ServerStreamingMethodResult<Response> = { [Symbol.asyncIterator](): AsyncIterator<Response, void> };
