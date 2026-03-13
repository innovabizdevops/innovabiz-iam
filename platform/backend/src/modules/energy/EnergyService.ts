import { Injectable } from "@nestjs/common";

@Injectable()
export class EnergyService {
  getHello(): string {
    return "Hello from Smart Energy Grid (EGY.01)!";
  }

  getCapabilities() {
    return undefined;
  }
}
