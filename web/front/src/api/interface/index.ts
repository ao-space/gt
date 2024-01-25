import { ClientConfig, ClientConfigBackend, transToFrontConfig } from "@/components/ClientConfigForm/interface";
import { ServerConfig } from "@/components/ServerConfigForm/interface";

// Request with no data
export interface Result {
  code: string;
  msg: string;
}

// Request with data
export interface ResultData<T = any> extends Result {
  data: T;
}

export namespace Login {
  export interface ReqLoginForm {
    username: string;
    password: string;
  }
  export interface ReqKeyValue {
    key: string;
  }
  export interface ResLogin {
    token: string;
  }
  export interface ResAuthButtons {
    [key: string]: string[];
  }
}
export namespace Register {
  export interface ReqRegisterForm {
    username: string;
    password: string;
    enablePprof: boolean;
  }
  export interface ResRegister {
    token: string;
  }
}

export namespace Server {
  export interface ResServerInfo {
    serverInfo: SystemState;
  }
  export interface SystemState {
    os: {
      goos: string;
      numCpu: number;
      compiler: string;
      goVersion: string;
      numGoroutine: number;
    };
    disk: {
      totalMb: number;
      usedMb: number;
      totalGb: number;
      usedGb: number;
      usedPercent: number;
    };
    cpu: {
      cores: number;
      cpus: number[];
    };
    ram: {
      totalMb: number;
      usedMb: number;
      usedPercent: number;
    };
  }
}

export namespace Config {
  export namespace Client {
    export interface ResConfig {
      config: ClientConfig.Config;
    }
    export interface ResConfigBackend {
      config: ClientConfigBackend.Config;
    }
    export function transClientConfigRes(config: ResConfigBackend): ResConfig {
      return {
        config: transToFrontConfig(config.config)
      };
    }
  }
  export namespace Server {
    export interface ResConfig {
      config: ServerConfig.Config;
    }
  }
}

export namespace Connection {
  export enum Status {
    Running = 0,
    Idle = 1,
    Wait = 2,
    Connecting = 3
  }
  export const StatusMap: { [key in Status]: string } = {
    [Status.Running]: "Running",
    [Status.Idle]: "Idle",
    [Status.Wait]: "Wait",
    [Status.Connecting]: "Connecting"
  };
  export interface LocalAddr {
    ip: string;
    port: number;
  }
  export interface RemoteAddr {
    ip: string;
    port: number;
  }
  export interface Connection {
    id?: string;
    family: number;
    type: number;
    localaddr: LocalAddr;
    remoteaddr: RemoteAddr;
    status: string;
  }
  export interface Pool {
    [key: string]: Status;
  }
  export interface ResConnection {
    external: Connection[];
    serverPool?: Connection[];
    clientPool?: Pool;
  }
}
