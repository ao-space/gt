import { ClientConfig } from "@/components/ClientConfigForm/interface";
// 请求响应参数（不包含data）
export interface Result {
  code: string;
  msg: string;
}

// 请求响应参数（包含data）
export interface ResultData<T = any> extends Result {
  data: T;
}

// 分页响应参数
export interface ResPage<T> {
  list: T[];
  pageNum: number;
  pageSize: number;
  total: number;
}

// 分页请求参数
export interface ReqPage {
  pageNum: number;
  pageSize: number;
}

// 文件上传模块
export namespace Upload {
  export interface ResFileUrl {
    fileUrl: string;
  }
}

// 登录模块
export namespace Login {
  export interface ReqLoginForm {
    username: string;
    password: string;
  }
  export interface ResLogin {
    token: string;
  }
  export interface ResAuthButtons {
    [key: string]: string[];
  }
}

// 用户管理模块
export namespace User {
  export interface ReqUserParams extends ReqPage {
    username: string;
    gender: number;
    idCard: string;
    email: string;
    address: string;
    createTime: string[];
    status: number;
  }
  export interface ResUserList {
    id: string;
    username: string;
    gender: number;
    user: { detail: { age: number } };
    idCard: string;
    email: string;
    address: string;
    createTime: string;
    status: number;
    avatar: string;
    photo: any[];
    children?: ResUserList[];
  }
  export interface ResStatus {
    userLabel: string;
    userValue: number;
  }
  export interface ResGender {
    genderLabel: string;
    genderValue: number;
  }
  export interface ResDepartment {
    id: string;
    name: string;
    children?: ResDepartment[];
  }
  export interface ResRole {
    id: string;
    name: string;
    children?: ResDepartment[];
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
interface ResClientConfig
  extends ClientConfig.GeneralSetting,
    ClientConfig.TCPForwardSetting,
    ClientConfig.LogSetting,
    ClientConfig.SentrySetting,
    ClientConfig.WebRTCSetting {
  Config: string;
  Services: ResClientServices[];
}
interface ResClientServices {
  HostPrefix: string;
  RemoteTCPPort: number;
  RemoteTCPRandom: boolean;
  LocalURL: string;
  LocalTimeout: string;
  UseLocalAsHTTPHost: boolean;
}

export namespace Config {
  export namespace Client {
    export interface ResConfig {
      config: ResClientConfig;
    }
  }
}
