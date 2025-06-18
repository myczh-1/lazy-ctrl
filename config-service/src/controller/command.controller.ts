import { Controller, Get, Post, Put, Del, Param, Body } from '@midwayjs/decorator';
import { CommandService } from '../service/command.service';
import { Inject } from '@midwayjs/decorator';

@Controller('/api/commands')
export class CommandController {
  @Inject()
  commandService: CommandService;

  @Get('/')
  async listCommands() {
    return await this.commandService.listCommands();
  }

  @Get('/:id')
  async getCommand(@Param('id') id: string) {
    return await this.commandService.getCommand(id);
  }

  @Post('/')
  async createCommand(@Body() command: any) {
    return await this.commandService.createCommand(command);
  }

  @Put('/:id')
  async updateCommand(@Param('id') id: string, @Body() command: any) {
    return await this.commandService.updateCommand(id, command);
  }

  @Del('/:id')
  async deleteCommand(@Param('id') id: string) {
    return await this.commandService.deleteCommand(id);
  }

  @Post('/:id/execute')
  async executeCommand(@Param('id') id: string, @Body() body: { args?: string[] }) {
    return await this.commandService.executeCommand(id, body.args || []);
  }
}