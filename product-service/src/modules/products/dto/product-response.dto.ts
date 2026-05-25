import { Product } from '../entities/product.entity';

export class ProductResponseDto {
  id!: string;
  name!: string;
  price!: number;
  qty!: number;
  createdAt!: Date;

  static fromEntity(entity: Product): ProductResponseDto {
    const dto = new ProductResponseDto();
    dto.id = entity.id;
    dto.name = entity.name;
    dto.price = Number(entity.price);
    dto.qty = entity.qty;
    dto.createdAt = entity.createdAt;
    return dto;
  }
}
