import { Injectable } from "@nestjs/common";

@Injectable()
export class AgroService {
  getHello(): string {
    return "Hello from Precision Agro (AGR.01)!";
  }

  getCapabilities() {
    return undefined;
  }
}
