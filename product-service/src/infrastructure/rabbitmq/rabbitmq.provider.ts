import { ConfigService } from '@nestjs/config';
import * as amqplib from 'amqplib';

export const RABBITMQ_CHANNEL = 'RABBITMQ_CHANNEL';
export const RABBITMQ_CONNECTION = 'RABBITMQ_CONNECTION';

export const RabbitMQProvider = {
  provide: RABBITMQ_CHANNEL,
  inject: [ConfigService],
  useFactory: async (config: ConfigService): Promise<amqplib.Channel> => {
    const url = config.get<string>('rabbitmq.url')!;

    let conn: Awaited<ReturnType<typeof amqplib.connect>>;
    let retries = 0;
    const maxRetries = 5;

    while (true) {
      try {
        conn = await amqplib.connect(url);
        break;
      } catch (err) {
        retries++;
        if (retries > maxRetries) throw err;
        const delay = retries * 1000;
        console.warn(
          `[RabbitMQ] Connection failed, retry ${retries}/${maxRetries} in ${delay}ms`,
        );
        await new Promise((r) => setTimeout(r, delay));
      }
    }

    conn.on('error', (err) =>
      console.error('[RabbitMQ] Connection error:', err.message),
    );
    conn.on('close', () =>
      console.warn(
        '[RabbitMQ] Connection closed — restart service to reconnect',
      ),
    );

    const channel = await conn.createChannel();
    channel.on('error', (err) =>
      console.error('[RabbitMQ] Channel error:', err.message),
    );

    console.log('[RabbitMQ] Channel ready');
    return channel;
  },
};
