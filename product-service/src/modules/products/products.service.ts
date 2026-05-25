import { Injectable, Logger, NotFoundException, Inject } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { CreateProductDto } from './dto/create-product.dto';
import { ProductResponseDto } from './dto/product-response.dto';
import { ProductsRepository } from './products.repository';
import { ProductPublisher } from './messaging/product.publisher';
import { Redis } from 'ioredis';
import { REDIS_CLIENT } from '../../infrastructure/redis/redis.provider';

@Injectable()
export class ProductsService {
  private readonly logger = new Logger(ProductsService.name);
  private readonly CACHE_TTL = 60;

  constructor(
    private readonly productsRepository: ProductsRepository,
    private readonly productPublisher: ProductPublisher,
    @Inject(REDIS_CLIENT) private readonly redis: Redis,
  ) {}

  async create(dto: CreateProductDto): Promise<ProductResponseDto> {
    const product = await this.productsRepository.create(dto);
    this.logger.log(`Product created: ${product.id}`);

    this.productPublisher.publish({
      productId: product.id,
      name: product.name,
      price: Number(product.price),
      qty: product.qty,
      createdAt: product.createdAt,
    });
    return ProductResponseDto.fromEntity(product);
  }

  async findById(id: string): Promise<ProductResponseDto> {
    const cacheKey = `product:${id}`;

    const cached = await this.redis.get(cacheKey);
    if (cached) {
      this.logger.log(`Cache hit: ${cacheKey}`);
      return JSON.parse(cached) as ProductResponseDto;
    }

    const product = await this.productsRepository.findById(id);
    if (!product) throw new NotFoundException(`Product ${id} not found`);

    const response = ProductResponseDto.fromEntity(product);
    await this.redis.setex(cacheKey, this.CACHE_TTL, JSON.stringify(response));
    this.logger.log(`Cache set: ${cacheKey}`);

    return response;
  }

  async handleOrderCreated(productId: string, quantity: number): Promise<void> {
    await this.productsRepository.decrementQty(productId, quantity);
    
    const updated = await this.productsRepository.findById(productId);
    if (updated) {
      const response = ProductResponseDto.fromEntity(updated);
      await this.redis.setex(
        `product:${productId}`,
        this.CACHE_TTL,
        JSON.stringify(response),
      );
      this.logger.log(`Cache refreshed: product:${productId}`);
    }
  }
}
