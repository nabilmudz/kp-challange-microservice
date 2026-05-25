import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { Product } from './entities/product.entity';
import { OrderConsumer } from './messaging/order.consumer';
import { ProductPublisher } from './messaging/product.publisher';
import { ProductsController } from './products.controller';
import { ProductsRepository } from './products.repository';
import { ProductsService } from './products.service';

@Module({
  imports: [TypeOrmModule.forFeature([Product])],
  controllers: [ProductsController],
  providers: [
    ProductsService,
    ProductsRepository,
    ProductPublisher,
    OrderConsumer,
  ],
  exports: [ProductsService],
})
export class ProductsModule {}
