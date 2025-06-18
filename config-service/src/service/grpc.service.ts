import { Provide, Init } from '@midwayjs/decorator';
import * as grpc from '@grpc/grpc-js';
import * as protoLoader from '@grpc/proto-loader';
import { join } from 'path';

@Provide()
export class GrpcService {
  private client: any;

  @Init()
  async init() {
    const PROTO_PATH = join(__dirname, '../../proto/controller.proto');
    const packageDefinition = protoLoader.loadSync(PROTO_PATH, {
      keepCase: true,
      longs: String,
      enums: String,
      defaults: true,
      oneofs: true,
    });

    const controllerProto = grpc.loadPackageDefinition(packageDefinition).controller as any;
    this.client = new controllerProto.ControllerService('localhost:50051', grpc.credentials.createInsecure());
  }

  async executeCommand(commandId: string, args: string[] = []): Promise<any> {
    return new Promise((resolve, reject) => {
      this.client.ExecuteCommand(
        {
          command_id: commandId,
          args: args,
        },
        (error: any, response: any) => {
          if (error) {
            reject(error);
          } else {
            resolve(response);
          }
        }
      );
    });
  }

  async listCommands(): Promise<any> {
    return new Promise((resolve, reject) => {
      this.client.ListCommands({}, (error: any, response: any) => {
        if (error) {
          reject(error);
        } else {
          resolve(response);
        }
      });
    });
  }
}