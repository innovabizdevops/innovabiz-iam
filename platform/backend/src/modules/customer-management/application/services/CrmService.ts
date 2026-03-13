import { Injectable } from "@nestjs/common";

@Injectable()
export class CrmService {
  constructor() {
    console.log("✅ [CustomerManagement] Service Initialized");
  }
}
