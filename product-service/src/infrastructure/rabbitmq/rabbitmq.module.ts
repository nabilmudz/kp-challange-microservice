import { Global, Module } from '@nestjs/common';
import { RabbitMQProvider } from './rabbitmq.provider';

@Global()
@Module({
  providers: [RabbitMQProvider],
  exports: [RabbitMQProvider],
})
export class RabbitMQModule {}
