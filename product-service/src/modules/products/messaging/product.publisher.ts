import { Inject, Injectable, Logger } from '@nestjs/common';
import * as amqplib from 'amqplib';
import { RABBITMQ_CHANNEL } from '../../../infrastructure/rabbitmq/rabbitmq.provider';

export const PRODUCT_CREATED_EXCHANGE = 'product.created';

@Injectable()
export class ProductPublisher {
  private readonly logger = new Logger(ProductPublisher.name);

  constructor(
    @Inject(RABBITMQ_CHANNEL)
    private readonly channel: amqplib.Channel,
  ) {}

  async onModuleInit() {
    await this.channel.assertExchange(PRODUCT_CREATED_EXCHANGE, 'fanout', {
      durable: true,
    });
    this.logger.log(`Publisher ready — exchange: ${PRODUCT_CREATED_EXCHANGE}`);
  }

  publish(payload: Record<string, unknown>): void {
    const content = Buffer.from(JSON.stringify(payload));
    this.channel.publish(PRODUCT_CREATED_EXCHANGE, '', content, {
      persistent: true,
    });
    this.logger.log(`Published: ${JSON.stringify(payload)}`);
  }
}