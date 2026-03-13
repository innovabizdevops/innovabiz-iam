import { Injectable } from "@nestjs/common";

@Injectable()
export class SwarmService {
  getHello(): string {
    return "Hello from Specialist Swarm (AX.04)!";
  }

  getCapabilities() {
    return undefined;
  }
}
