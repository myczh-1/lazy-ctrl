import { Provide } from '@midwayjs/decorator';
import { GrpcService } from './grpc.service';
import { ConfigService } from './config.service';
import { Inject } from '@midwayjs/decorator';

export interface Command {
  id: string;
  name: string;
  description: string;
  script_path: string;
  args?: string[];
  env?: Record<string, string>;
  work_dir?: string;
}

@Provide()
export class CommandService {
  @Inject()
  grpcService: GrpcService;

  @Inject()
  configService: ConfigService;

  async listCommands(): Promise<Command[]> {
    const config = await this.configService.getConfig();
    return config.commands;
  }

  async getCommand(id: string): Promise<Command | null> {
    const config = await this.configService.getConfig();
    return config.commands.find(cmd => cmd.id === id) || null;
  }

  async createCommand(command: Command): Promise<Command> {
    const config = await this.configService.getConfig();
    config.commands.push(command);
    await this.configService.saveConfig(config);
    return command;
  }

  async updateCommand(id: string, updatedCommand: Partial<Command>): Promise<Command | null> {
    const config = await this.configService.getConfig();
    const index = config.commands.findIndex(cmd => cmd.id === id);
    
    if (index === -1) {
      return null;
    }

    config.commands[index] = { ...config.commands[index], ...updatedCommand };
    await this.configService.saveConfig(config);
    return config.commands[index];
  }

  async deleteCommand(id: string): Promise<boolean> {
    const config = await this.configService.getConfig();
    const index = config.commands.findIndex(cmd => cmd.id === id);
    
    if (index === -1) {
      return false;
    }

    config.commands.splice(index, 1);
    await this.configService.saveConfig(config);
    return true;
  }

  async executeCommand(id: string, args: string[] = []): Promise<any> {
    return await this.grpcService.executeCommand(id, args);
  }
}