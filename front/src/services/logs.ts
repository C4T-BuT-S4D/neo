import { grpcAddress } from "@/config.ts";
import { ServiceClient, ServiceDefinition } from "@/proto/logs/api.ts";
import { useAuthContext } from "@/routes/authContext";
import {
  WebsocketTransport,
  createChannel,
  createClientFactory,
} from "nice-grpc-web";
import { useMemo } from "react";

const channel = createChannel(grpcAddress, WebsocketTransport());

export type LogsServiceClient = ServiceClient;

export function useLogsServiceClient(): LogsServiceClient {
  const authContext = useAuthContext();
  return useMemo(
    () =>
      createClientFactory()
        .use((call, options) =>
          call.next(call.request, {
            ...options,
            metadata: {
              ...options.metadata,
              ...authContext.metadata,
            },
          })
        )
        .create(ServiceDefinition, channel),
    [authContext]
  );
}
