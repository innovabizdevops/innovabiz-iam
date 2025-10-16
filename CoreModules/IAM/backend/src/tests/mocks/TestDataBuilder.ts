/**
 * 🏗️ TEST DATA BUILDER - IAM MODULE
 * Builder para criação de dados de teste padronizados
 * Versão: 2.1.0 | Data: 2025-01-27 | Autor: Eduardo Jeremias
 * 
 * Padrão: Builder Pattern para testes
 * Objetivo: Facilitar criação de dados consistentes para testes
 */

import { Repository } from 'typeorm';
import * as bcrypt from 'bcrypt';

// Entidades
import { User } from '../../entities/User';
import { Role } from '../../entities/Role';
import { Permission } from '../../entities/Permission';
import { Session } from '../../entities/Session';

/**
 * Interface para dados de usuário de teste
 */
interface TestUserData {
  email: string;
  username: string;
  password: string;
  firstName: string;
  lastName: string;
  roles: Role[];
  isActive: boolean;
  isEmailVerified: boolean;
  tenantId: string;
  phone?: string;
  dateOfBirth?: Date;
  profilePicture?: string;
}

/**
 * Interface para dados de role de teste
 */
interface TestRoleData {
  name: string;
  description: string;
  permissions: Permission[];
  isActive?: boolean;
}

/**
 * Interface para dados de permissão de teste
 */
interface TestPermissionData {
  name: string;
  description: string;
  resource?: string;
  action?: string;
}

/**
 * Interface para dados de sessão de teste
 */
interface TestSessionData {
  userId: string;
  tenantId: string;
  ipAddress: string;
  userAgent: string;
  isActive?: boolean;
  expiresAt?: Date;
}

/**
 * Builder para criação de dados de teste
 */
export class TestDataBuilder {
  constructor(
    private userRepository: Repository<User>,
    private roleRepository: Repository<Role>,
    private permissionRepository: Repository<Permission>,
    private sessionRepository?: Repository<Session>
  ) {}

  /**
   * Cria usuário de teste com dados padrão
   */
  async createUser(userData: Partial<TestUserData> = {}): Promise<User> {
    const defaultData: TestUserData = {
      email: 'test@innovabiz.com',
      username: 'testuser',
      password: 'TestPassword123!',
      firstName: 'Test',
      lastName: 'User',
      roles: [],
      isActive: true,
      isEmailVerified: true,
      tenantId: 'tenant-1',
      ...userData
    };

    // Hash da senha
    const hashedPassword = await bcrypt.hash(defaultData.password, 12);

    const user = this.userRepository.create({
      ...defaultData,
      password: hashedPassword,
      createdAt: new Date(),
      updatedAt: new Date()
    });

    return await this.userRepository.save(user);
  }

  /**
   * Cria role de teste com dados padrão
   */
  async createRole(roleData: Partial<TestRoleData> = {}): Promise<Role> {
    const defaultData: TestRoleData = {
      name: 'test-role',
      description: 'Role de teste',
      permissions: [],
      isActive: true,
      ...roleData
    };

    const role = this.roleRepository.create({
      ...defaultData,
      createdAt: new Date(),
      updatedAt: new Date()
    });

    return await this.roleRepository.save(role);
  }

  /**
   * Cria permissão de teste com dados padrão
   */
  async createPermission(permissionData: Partial<TestPermissionData> = {}): Promise<Permission> {
    const defaultData: TestPermissionData = {
      name: 'test:permission',
      description: 'Permissão de teste',
      resource: 'test',
      action: 'read',
      ...permissionData
    };

    const permission = this.permissionRepository.create({
      ...defaultData,
      createdAt: new Date(),
      updatedAt: new Date()
    });

    return await this.permissionRepository.save(permission);
  }

  /**
   * Cria múltiplas permissões de teste
   */
  async createPermissions(permissionsData: Partial<TestPermissionData>[]): Promise<Permission[]> {
    const permissions: Permission[] = [];
    
    for (const permissionData of permissionsData) {
      const permission = await this.createPermission(permissionData);
      permissions.push(permission);
    }

    return permissions;
  }

  /**
   * Cria sessão de teste com dados padrão
   */
  async createSession(sessionData: Partial<TestSessionData> = {}): Promise<Session> {
    if (!this.sessionRepository) {
      throw new Error('Session repository não fornecido');
    }

    const defaultData: TestSessionData = {
      userId: '1',
      tenantId: 'tenant-1',
      ipAddress: '192.168.1.1',
      userAgent: 'Test-Agent',
      isActive: true,
      expiresAt: new Date(Date.now() + 24 * 60 * 60 * 1000), // 24 horas
      ...sessionData
    };

    const session = this.sessionRepository.create({
      ...defaultData,
      createdAt: new Date(),
      updatedAt: new Date()
    });

    return await this.sessionRepository.save(session);
  }

  /**
   * Cria usuário administrador de teste
   */
  async createAdminUser(tenantId: string = 'tenant-1'): Promise<User> {
    // Criar permissões administrativas
    const adminPermissions = await this.createPermissions([
      { name: 'user:read', description: 'Ler usuários' },
      { name: 'user:write', description: 'Escrever usuários' },
      { name: 'user:delete', description: 'Deletar usuários' },
      { name: 'role:read', description: 'Ler roles' },
      { name: 'role:write', description: 'Escrever roles' },
      { name: 'role:delete', description: 'Deletar roles' },
      { name: 'admin:all', description: 'Acesso administrativo completo' }
    ]);

    // Criar role de administrador
    const adminRole = await this.createRole({
      name: 'admin',
      description: 'Administrador do sistema',
      permissions: adminPermissions
    });

    // Criar usuário administrador
    return await this.createUser({
      email: 'admin@innovabiz.com',
      username: 'admin',
      password: 'AdminPassword123!',
      firstName: 'Admin',
      lastName: 'User',
      roles: [adminRole],
      isActive: true,
      isEmailVerified: true,
      tenantId
    });
  }

  /**
   * Cria usuário regular de teste
   */
  async createRegularUser(tenantId: string = 'tenant-1'): Promise<User> {
    // Criar permissões básicas
    const userPermissions = await this.createPermissions([
      { name: 'profile:read', description: 'Ler próprio perfil' },
      { name: 'profile:write', description: 'Editar próprio perfil' }
    ]);

    // Criar role de usuário
    const userRole = await this.createRole({
      name: 'user',
      description: 'Usuário padrão',
      permissions: userPermissions
    });

    // Criar usuário regular
    return await this.createUser({
      email: 'user@innovabiz.com',
      username: 'user',
      password: 'UserPassword123!',
      firstName: 'Regular',
      lastName: 'User',
      roles: [userRole],
      isActive: true,
      isEmailVerified: true,
      tenantId
    });
  }

  /**
   * Cria usuário inativo de teste
   */
  async createInactiveUser(tenantId: string = 'tenant-1'): Promise<User> {
    return await this.createUser({
      email: 'inactive@innovabiz.com',
      username: 'inactive',
      password: 'InactivePassword123!',
      firstName: 'Inactive',
      lastName: 'User',
      roles: [],
      isActive: false,
      isEmailVerified: true,
      tenantId
    });
  }

  /**
   * Cria usuário com email não verificado
   */
  async createUnverifiedUser(tenantId: string = 'tenant-1'): Promise<User> {
    return await this.createUser({
      email: 'unverified@innovabiz.com',
      username: 'unverified',
      password: 'UnverifiedPassword123!',
      firstName: 'Unverified',
      lastName: 'User',
      roles: [],
      isActive: true,
      isEmailVerified: false,
      tenantId
    });
  }

  /**
   * Cria usuário bloqueado (com tentativas de login excedidas)
   */
  async createLockedUser(tenantId: string = 'tenant-1'): Promise<User> {
    const user = await this.createUser({
      email: 'locked@innovabiz.com',
      username: 'locked',
      password: 'LockedPassword123!',
      firstName: 'Locked',
      lastName: 'User',
      roles: [],
      isActive: true,
      isEmailVerified: true,
      tenantId
    });

    // Simular bloqueio por tentativas excessivas
    await this.userRepository.update(user.id, {
      loginAttempts: 5,
      lockedUntil: new Date(Date.now() + 15 * 60 * 1000) // 15 minutos
    });

    return user;
  }

  /**
   * Cria conjunto completo de dados de teste
   */
  async createTestDataSet(tenantId: string = 'tenant-1'): Promise<{
    adminUser: User;
    regularUser: User;
    inactiveUser: User;
    unverifiedUser: User;
    lockedUser: User;
    adminRole: Role;
    userRole: Role;
    permissions: Permission[];
  }> {
    // Criar permissões
    const permissions = await this.createPermissions([
      { name: 'user:read', description: 'Ler usuários' },
      { name: 'user:write', description: 'Escrever usuários' },
      { name: 'user:delete', description: 'Deletar usuários' },
      { name: 'role:read', description: 'Ler roles' },
      { name: 'role:write', description: 'Escrever roles' },
      { name: 'profile:read', description: 'Ler próprio perfil' },
      { name: 'profile:write', description: 'Editar próprio perfil' },
      { name: 'admin:all', description: 'Acesso administrativo completo' }
    ]);

    // Criar roles
    const adminRole = await this.createRole({
      name: 'admin',
      description: 'Administrador do sistema',
      permissions: permissions
    });

    const userRole = await this.createRole({
      name: 'user',
      description: 'Usuário padrão',
      permissions: permissions.slice(5, 7) // Apenas permissões de perfil
    });

    // Criar usuários
    const adminUser = await this.createUser({
      email: 'admin@innovabiz.com',
      username: 'admin',
      password: 'AdminPassword123!',
      firstName: 'Admin',
      lastName: 'User',
      roles: [adminRole],
      isActive: true,
      isEmailVerified: true,
      tenantId
    });

    const regularUser = await this.createUser({
      email: 'user@innovabiz.com',
      username: 'user',
      password: 'UserPassword123!',
      firstName: 'Regular',
      lastName: 'User',
      roles: [userRole],
      isActive: true,
      isEmailVerified: true,
      tenantId
    });

    const inactiveUser = await this.createInactiveUser(tenantId);
    const unverifiedUser = await this.createUnverifiedUser(tenantId);
    const lockedUser = await this.createLockedUser(tenantId);

    return {
      adminUser,
      regularUser,
      inactiveUser,
      unverifiedUser,
      lockedUser,
      adminRole,
      userRole,
      permissions
    };
  }

  /**
   * Limpa todos os dados de teste
   */
  async cleanupTestData(): Promise<void> {
    // Ordem importante devido às foreign keys
    if (this.sessionRepository) {
      await this.sessionRepository.clear();
    }
    await this.userRepository.clear();
    await this.roleRepository.clear();
    await this.permissionRepository.clear();
  }

  /**
   * Cria dados de teste para multi-tenancy
   */
  async createMultiTenantTestData(): Promise<{
    tenant1Users: User[];
    tenant2Users: User[];
    sharedRoles: Role[];
  }> {
    // Criar roles compartilhadas
    const sharedPermissions = await this.createPermissions([
      { name: 'read:basic', description: 'Leitura básica' },
      { name: 'write:basic', description: 'Escrita básica' }
    ]);

    const sharedRoles = [
      await this.createRole({
        name: 'basic-user',
        description: 'Usuário básico',
        permissions: sharedPermissions
      })
    ];

    // Criar usuários para tenant-1
    const tenant1Users = [
      await this.createUser({
        email: 'user1@tenant1.com',
        username: 'user1-t1',
        tenantId: 'tenant-1',
        roles: sharedRoles
      }),
      await this.createUser({
        email: 'user2@tenant1.com',
        username: 'user2-t1',
        tenantId: 'tenant-1',
        roles: sharedRoles
      })
    ];

    // Criar usuários para tenant-2
    const tenant2Users = [
      await this.createUser({
        email: 'user1@tenant2.com',
        username: 'user1-t2',
        tenantId: 'tenant-2',
        roles: sharedRoles
      }),
      await this.createUser({
        email: 'user2@tenant2.com',
        username: 'user2-t2',
        tenantId: 'tenant-2',
        roles: sharedRoles
      })
    ];

    return {
      tenant1Users,
      tenant2Users,
      sharedRoles
    };
  }
}