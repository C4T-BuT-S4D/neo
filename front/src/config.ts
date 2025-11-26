let grpcAddressTemp = "ws://127.0.0.1:5005";
if (!import.meta.env.DEV) {
  if (window.location.protocol === "https:") {
    grpcAddressTemp = "wss://" + window.location.host;
  } else {
    grpcAddressTemp = "ws://" + window.location.host;
  }
}

export const grpcAddress = grpcAddressTemp;
