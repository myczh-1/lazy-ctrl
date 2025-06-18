import { Provide } from '@midwayjs/decorator';
import { readFile, writeFile } from 'fs/promises';
import { join } from 'path';

export interface SecurityConfig {
  allowed_paths: string[];
  whitelist: string[];
  require_auth: boolean;
}

export interface Config {
  commands: Array<{
    id: string;
    name: string;
    description: string;
    script_path: string;
    args?: string[];
    env?: Record<string, string>;
    work_dir?: string;
  }>;
  security: SecurityConfig;
}

@Provide()
export class ConfigService {
  private configPath = join(__dirname, '../../data/commands.json');

  async getConfig(): Promise<Config> {
    try {
      const data = await readFile(this.configPath, 'utf-8');
      return JSON.parse(data);
    } catch (error) {
      // Return default config if file doesn't exist
      return {
        commands: [],
        security: {
          allowed_paths: [],
          whitelist: [],
          require_auth: false,
        },
      };
    }
  }

  async saveConfig(config: Config): Promise<void> {
    const data = JSON.stringify(config, null, 2);
    await writeFile(this.configPath, data, 'utf-8');
  }
}