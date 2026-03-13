import { Injectable } from "@nestjs/common";
import { Account } from "../../domain/entities/Account";
import { CreateAccountDto } from "../dto/CreateAccountDto";
import { Money } from "../../domain/entities/Money";

@Injectable()
export class FinService {
  private readonly accounts: Map<string, Account> = new Map(); // Mock Repository

  constructor() {
    console.log("✅ [FinancialManagement] Service Initialized w/ BIAN Logic");
  }

  createAccount(dto: CreateAccountDto): Account {
    const id = Math.random().toString(36).substring(7); // Mock ID
    const account = Account.create(id, dto.customerId, dto.type, dto.currency);

    if (dto.initialDeposit && dto.initialDeposit > 0) {
      account.deposit(new Money(dto.initialDeposit, dto.currency));
    }

    this.accounts.set(id, account);
    console.log(
      `🏦 [BIAN] Created Account ${id} for Customer ${dto.customerId}`,
    );
    return account;
  }

  getAccount(id: string): Account | undefined {
    return this.accounts.get(id);
  }
}
