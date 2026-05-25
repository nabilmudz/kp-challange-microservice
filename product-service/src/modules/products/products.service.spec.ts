import { NotFoundException } from '@nestjs/common';
import { Test, TestingModule } from '@nestjs/testing';
import { REDIS_CLIENT } from '../../infrastructure/redis/redis.provider';
import { CreateProductDto } from './dto/create-product.dto';
import { Product } from './entities/product.entity';
import { ProductPublisher } from './messaging/product.publisher';
import { ProductsRepository } from './products.repository';
import { ProductsService } from './products.service';

const mockProduct = (): Product => ({
  id: 'uuid-123',
  name: 'Mechanical Keyboard',
  price: 850000,
  qty: 50,
  createdAt: new Date('2026-01-01'),
});

const mockRedis = {
  get: jest.fn(),
  setex: jest.fn(),
  del: jest.fn(),
};

const mockRepository = {
  create: jest.fn(),
  findById: jest.fn(),
  decrementQty: jest.fn(),
};

const mockPublisher = {
  publish: jest.fn(),
};

describe('ProductsService', () => {
  let service: ProductsService;

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      providers: [
        ProductsService,
        { provide: ProductsRepository, useValue: mockRepository },
        { provide: ProductPublisher, useValue: mockPublisher },
        { provide: REDIS_CLIENT, useValue: mockRedis },
      ],
    }).compile();

    service = module.get(ProductsService);
  });

  afterEach(() => jest.clearAllMocks());

  describe('create', () => {
    it('should create product, publish event, and return response dto', async () => {
      const dto: CreateProductDto = {
        name: 'Mechanical Keyboard',
        price: 850000,
        qty: 50,
      };
      mockRepository.create.mockResolvedValue(mockProduct());

      const result = await service.create(dto);

      expect(mockRepository.create).toHaveBeenCalledWith(dto);
      expect(mockPublisher.publish).toHaveBeenCalledWith(
        expect.objectContaining({ productId: 'uuid-123' }),
      );
      expect(result.id).toBe('uuid-123');
      expect(result.price).toBe(850000);
    });

    it('should throw if repository fails', async () => {
      mockRepository.create.mockRejectedValue(new Error('DB error'));

      await expect(
        service.create({ name: 'X', price: 1000, qty: 1 }),
      ).rejects.toThrow('DB error');

      expect(mockPublisher.publish).not.toHaveBeenCalled();
    });
  });

  describe('findById', () => {
    it('should return cached product on cache hit', async () => {
      mockRedis.get.mockResolvedValue(
        JSON.stringify({
          id: 'uuid-123',
          name: 'Mechanical Keyboard',
          price: 850000,
          qty: 50,
          createdAt: '2026-01-01T00:00:00.000Z',
        }),
      );

      const result = await service.findById('uuid-123');

      expect(mockRedis.get).toHaveBeenCalledWith('product:uuid-123');
      expect(mockRepository.findById).not.toHaveBeenCalled();
      expect(result.id).toBe('uuid-123');
    });

    it('should hit DB and set cache on cache miss', async () => {
      mockRedis.get.mockResolvedValue(null);
      mockRepository.findById.mockResolvedValue(mockProduct());
      mockRedis.setex.mockResolvedValue('OK');

      const result = await service.findById('uuid-123');

      expect(mockRepository.findById).toHaveBeenCalledWith('uuid-123');
      expect(mockRedis.setex).toHaveBeenCalledWith(
        'product:uuid-123',
        60,
        expect.any(String),
      );
      expect(result.id).toBe('uuid-123');
    });

    it('should throw NotFoundException when product not found in DB', async () => {
      mockRedis.get.mockResolvedValue(null);
      mockRepository.findById.mockResolvedValue(null);

      await expect(service.findById('uuid-999')).rejects.toThrow(
        NotFoundException,
      );
      expect(mockRedis.setex).not.toHaveBeenCalled();
    });
  });

  describe('handleOrderCreated', () => {
    it('should decrement qty and refresh cache on success', async () => {
      mockRepository.decrementQty.mockResolvedValue(true);
      mockRepository.findById.mockResolvedValue(mockProduct());
      mockRedis.setex.mockResolvedValue('OK');

      await service.handleOrderCreated('uuid-123', 2);

      expect(mockRepository.decrementQty).toHaveBeenCalledWith('uuid-123', 2);
      expect(mockRepository.findById).toHaveBeenCalledWith('uuid-123');
      expect(mockRedis.setex).toHaveBeenCalledWith(
        'product:uuid-123',
        60,
        expect.any(String),
      );
    });

    it('should not refresh cache when product not found after decrement', async () => {
      mockRepository.decrementQty.mockResolvedValue(true);
      mockRepository.findById.mockResolvedValue(null);

      await service.handleOrderCreated('uuid-123', 2);

      expect(mockRedis.setex).not.toHaveBeenCalled();
    });
  });
});
