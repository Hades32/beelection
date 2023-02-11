# 🐝-lection

## Fly Env Vars

```shell
# mine 
FLY_API_TOKEN=xxxx
FLY_API_HOSTNAME="_api.internal:4280"

# official
FLY_PUBLIC_IP=2604:1380:4091:360d:0:5ce9:9ed6:1
FLY_VM_MEMORY_MB=256
FLY_IMAGE_REF=registry.fly.io/...
FLY_ALLOC_ID=e784eeeea29e68
FLY_REGION=fra
FLY_APP_NAME=beelection
```


export async function stopMachine(id: string, app: string) {
  const resp = await fetch(
    `http://${env.FLY_API_HOSTNAME}/v1/apps/${app}/machines/${id}/stop`,
    {
      headers: {
        Authorization: `Bearer ${env.FLY_API_TOKEN}`,
        "Content-Type": "application/json",
      },
      method: "POST",
    }
  );
  if (resp.status !== 200) {
    throw new BadStatusError(resp.status);
  }
}

export async function deleteMachine(id: string, app: string) {
  const resp = await fetch(
    `http://${env.FLY_API_HOSTNAME}/v1/apps/${app}/machines/${id}`,
    {
      headers: {
        Authorization: `Bearer ${env.FLY_API_TOKEN}`,
        "Content-Type": "application/json",
      },
      method: "DELETE",
    }
  );
  if (resp.status !== 200) {
    throw new BadStatusError(resp.status);
  }
}

export async function createMachine(
  app: string,
  config: {
    name?: string;
    config: {
      image: string;
      env?: Record<string, string>;
      services?: Array<{
        ports: Array<{
          port: number;
          handlers: Array<"tls" | "http">;
        }>;
        protocol: "tcp" | "udp";
        internal_port: number;
      }>;
      checks?: {
        httpget?: {
          type: string;
          port: number;
          method: string;
          path: string;
          interval: string;
          timeout: string;
        };
      };
    };
  }
) {
  const resp = await fetch(
    `http://${env.FLY_API_HOSTNAME}/v1/apps/${app}/machines`,
    {
      headers: {
        Authorization: `Bearer ${env.FLY_API_TOKEN}`,
        "Content-Type": "application/json",
      },
      method: "POST",
      body: JSON.stringify(config),
    }
  );
  if (resp.status !== 200) {
    throw new BadStatusError(resp.status);
  }
}

