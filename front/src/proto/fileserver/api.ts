// @ts-nocheck
/* eslint-disable */
import Long from "long";
import type { CallContext, CallOptions } from "nice-grpc-common";
import _m0 from "protobufjs/minimal";

export const protobufPackage = "fileserver";

export interface FileInfo {
  uuid: string;
}

export interface FileStream {
  chunk: Uint8Array;
}

function createBaseFileInfo(): FileInfo {
  return { uuid: "" };
}

export const FileInfo = {
  encode(message: FileInfo, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.uuid !== "") {
      writer.uint32(10).string(message.uuid);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): FileInfo {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseFileInfo();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.uuid = reader.string();
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): FileInfo {
    return { uuid: isSet(object.uuid) ? globalThis.String(object.uuid) : "" };
  },

  toJSON(message: FileInfo): unknown {
    const obj: any = {};
    if (message.uuid !== "") {
      obj.uuid = message.uuid;
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<FileInfo>, I>>(base?: I): FileInfo {
    return FileInfo.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<FileInfo>, I>>(object: I): FileInfo {
    const message = createBaseFileInfo();
    message.uuid = object.uuid ?? "";
    return message;
  },
};

function createBaseFileStream(): FileStream {
  return { chunk: new Uint8Array(0) };
}

export const FileStream = {
  encode(message: FileStream, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.chunk.length !== 0) {
      writer.uint32(10).bytes(message.chunk);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): FileStream {
    const reader = input instanceof _m0.Reader ? input : _m0.Reader.create(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseFileStream();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.chunk = reader.bytes();
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skipType(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): FileStream {
    return { chunk: isSet(object.chunk) ? bytesFromBase64(object.chunk) : new Uint8Array(0) };
  },

  toJSON(message: FileStream): unknown {
    const obj: any = {};
    if (message.chunk.length !== 0) {
      obj.chunk = base64FromBytes(message.chunk);
    }
    return obj;
  },

  create<I extends Exact<DeepPartial<FileStream>, I>>(base?: I): FileStream {
    return FileStream.fromPartial(base ?? ({} as any));
  },
  fromPartial<I extends Exact<DeepPartial<FileStream>, I>>(object: I): FileStream {
    const message = createBaseFileStream();
    message.chunk = object.chunk ?? new Uint8Array(0);
    return message;
  },
};

export type ServiceDefinition = typeof ServiceDefinition;
export const ServiceDefinition = {
  name: "Service",
  fullName: "fileserver.Service",
  methods: {
    uploadFile: {
      name: "UploadFile",
      requestType: FileStream,
      requestStream: true,
      responseType: FileInfo,
      responseStream: false,
      options: {},
    },
    downloadFile: {
      name: "DownloadFile",
      requestType: FileInfo,
      requestStream: false,
      responseType: FileStream,
      responseStream: true,
      options: {},
    },
  },
} as const;

export interface ServiceImplementation<CallContextExt = {}> {
  uploadFile(request: AsyncIterable<FileStream>, context: CallContext & CallContextExt): Promise<DeepPartial<FileInfo>>;
  downloadFile(
    request: FileInfo,
    context: CallContext & CallContextExt,
  ): ServerStreamingMethodResult<DeepPartial<FileStream>>;
}

export interface ServiceClient<CallOptionsExt = {}> {
  uploadFile(
    request: AsyncIterable<DeepPartial<FileStream>>,
    options?: CallOptions & CallOptionsExt,
  ): Promise<FileInfo>;
  downloadFile(request: DeepPartial<FileInfo>, options?: CallOptions & CallOptionsExt): AsyncIterable<FileStream>;
}

function bytesFromBase64(b64: string): Uint8Array {
  if (globalThis.Buffer) {
    return Uint8Array.from(globalThis.Buffer.from(b64, "base64"));
  } else {
    const bin = globalThis.atob(b64);
    const arr = new Uint8Array(bin.length);
    for (let i = 0; i < bin.length; ++i) {
      arr[i] = bin.charCodeAt(i);
    }
    return arr;
  }
}

function base64FromBytes(arr: Uint8Array): string {
  if (globalThis.Buffer) {
    return globalThis.Buffer.from(arr).toString("base64");
  } else {
    const bin: string[] = [];
    arr.forEach((byte) => {
      bin.push(globalThis.String.fromCharCode(byte));
    });
    return globalThis.btoa(bin.join(""));
  }
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
