import { Injectable } from "@nestjs/common";

@Injectable()
export class CompService {
  constructor() {
    console.log("✅ [ComplianceManagement] Service Initialized");
  }
}
