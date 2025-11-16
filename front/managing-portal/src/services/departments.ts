const API_BASE_URL = import.meta.env.VITE_API_URL || '';

// Department interfaces
export interface Department {
  id: string;
  name: string;
  description: string;
  parent_id?: string;
  organization_id?: string;
  level: number;
  path: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface DepartmentTreeNode {
  id: string;
  name: string;
  description: string;
  parent_id?: string;
  organization_id?: string;
  level: number;
  path: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  children?: DepartmentTreeNode[];
}

export interface DepartmentWithStats extends Department {
  user_count: number;
  child_count: number;
  total_users_count: number;
}

export interface CreateDepartmentRequest {
  name: string;
  description?: string;
  parent_id?: string;
  organization_id?: string;
}

export interface UpdateDepartmentRequest {
  name?: string;
  description?: string;
  parent_id?: string;
  organization_id?: string;
  is_active?: boolean;
}

const getAuthHeaders = () => {
  const token = localStorage.getItem('token') || sessionStorage.getItem('token');
  return {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`,
  };
};

export const departmentsApi = {
  /**
   * Get list of departments (flat list)
   */
  async getDepartments(parentId?: string, includeAll: boolean = false): Promise<Department[]> {
    const params = new URLSearchParams();
    if (parentId) params.append('parent_id', parentId);
    if (includeAll) params.append('include_all', 'true');

    const response = await fetch(`${API_BASE_URL}/api/v1/departments?${params}`, {
      headers: getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error(`Failed to fetch departments: ${response.statusText}`);
    }

    const data = await response.json();
    // API returns {items: [], total: n, ...} structure
    return data.items || [];
  },

  /**
   * Get departments as hierarchical tree
   */
  async getDepartmentTree(rootId?: string): Promise<DepartmentTreeNode> {
    const params = new URLSearchParams();
    params.append('tree', 'true');
    if (rootId) params.append('parent_id', rootId);

    const response = await fetch(`${API_BASE_URL}/api/v1/departments?${params}`, {
      headers: getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error(`Failed to fetch department tree: ${response.statusText}`);
    }

    return response.json();
  },

  /**
   * Get department by ID
   */
  async getDepartment(id: string, includeStats: boolean = false): Promise<Department | DepartmentWithStats> {
    const params = new URLSearchParams();
    if (includeStats) params.append('stats', 'true');

    const response = await fetch(`${API_BASE_URL}/api/v1/departments/${id}?${params}`, {
      headers: getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error(`Failed to fetch department: ${response.statusText}`);
    }

    return response.json();
  },

  /**
   * Create new department
   */
  async createDepartment(data: CreateDepartmentRequest): Promise<Department> {
    const response = await fetch(`${API_BASE_URL}/api/v1/departments`, {
      method: 'POST',
      headers: getAuthHeaders(),
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Failed to create department');
    }

    return response.json();
  },

  /**
   * Update department
   */
  async updateDepartment(id: string, data: UpdateDepartmentRequest): Promise<Department> {
    const response = await fetch(`${API_BASE_URL}/api/v1/departments/${id}`, {
      method: 'PUT',
      headers: getAuthHeaders(),
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Failed to update department');
    }

    return response.json();
  },

  /**
   * Delete department (soft delete)
   */
  async deleteDepartment(id: string): Promise<void> {
    const response = await fetch(`${API_BASE_URL}/api/v1/departments/${id}`, {
      method: 'DELETE',
      headers: getAuthHeaders(),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Failed to delete department');
    }
  },

  /**
   * Get child departments
   */
  async getChildren(id: string): Promise<Department[]> {
    const response = await fetch(`${API_BASE_URL}/api/v1/departments/${id}/children`, {
      headers: getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error(`Failed to fetch child departments: ${response.statusText}`);
    }

    return response.json();
  },
};
