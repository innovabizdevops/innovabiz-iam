import { Injectable } from "@nestjs/common";

@Injectable()
export class SimulationService {
  getHello(): string {
    return "Hello from Scenario Simulator (STR.01)!";
  }

  getCapabilities() {
    return undefined;
  }
}
