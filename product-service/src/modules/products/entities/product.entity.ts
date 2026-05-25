import {
  Column,
  CreateDateColumn,
  Entity,
  PrimaryGeneratedColumn,
} from 'typeorm';

@Entity('products')
export class Product {
  @PrimaryGeneratedColumn('uuid')
  id!: string;

  @Column({ type: 'varchar', nullable: false })
  name!: string;

  @Column({ type: 'decimal', precision: 12, scale: 2, nullable: false })
  price!: number;

  @Column({ type: 'int', nullable: false, default: 0 })
  qty!: number;

  @CreateDateColumn({ name: 'created_at' })
  createdAt!: Date;
}
