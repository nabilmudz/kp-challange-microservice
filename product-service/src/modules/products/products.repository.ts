import { Injectable } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { Product } from './entities/product.entity';

@Injectable()
export class ProductsRepository {
  constructor(
    @InjectRepository(Product)
    private readonly repo: Repository<Product>,
  ) {}

  async create(data: Partial<Product>): Promise<Product> {
    const product = this.repo.create(data);
    return this.repo.save(product);
  }

  async findById(id: string): Promise<Product | null> {
    return this.repo.findOne({ where: { id } });
  }

  async decrementQty(id: string, qty: number): Promise<boolean> {
    const result = await this.repo
      .createQueryBuilder()
      .update(Product)
      .set({ qty: () => `qty - ${qty}` })
      .where('id = :id AND qty >= :qty', { id, qty })
      .execute();

    return (result.affected ?? 0) > 0;
  }
}
