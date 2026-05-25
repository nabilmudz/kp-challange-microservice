import { Inject, Injectable, Logger, OnModuleInit } from '@nestjs/common';
import * as amqplib from 'amqplib';
import { RABBITMQ_CHANNEL } from '../../../infrastructure/rabbitmq/rabbitmq.provider';
import { ProductsService } from '../products.service';

export const ORDER_CREATED_EXCHANGE = 'order.created';

interface OrderCreatedPayload {
  orderId: string;
  productId: string;
  quantity: number;
}

@Injectable()
export class OrderConsumer implements OnModuleInit {
  private readonly logger = new Logger(OrderConsumer.name);

  constructor(
    @Inject(RABBITMQ_CHANNEL)
    private readonly channel: amqplib.Channel,
    private readonly productsService: ProductsService,
  ) {}

  async onModuleInit() {
    await this.channel.assertExchange(ORDER_CREATED_EXCHANGE, 'fanout', {
      durable: true,
    });

    const q = await this.channel.assertQueue(
      'product-service.order.created',
      { durable: true },
    );

    await this.channel.bindQueue(q.queue, ORDER_CREATED_EXCHANGE, '');

    this.channel.prefetch(10);
    this.logger.log(`Consumer ready — listening on exchange: ${ORDER_CREATED_EXCHANGE}`);

    this.channel.consume(q.queue, async (msg) => {
      if (!msg) return;
      try {
        const payload = JSON.parse(msg.content.toString()) as OrderCreatedPayload;
        this.logger.log(`Received order.created: ${JSON.stringify(payload)}`);
        await this.productsService.handleOrderCreated(
          payload.productId,
          payload.quantity,
        );
        this.channel.ack(msg);
      } catch (err) {
        this.logger.error('Failed to process order.created', err);
        this.channel.nack(msg, false, false);
      }
    });
  }
}