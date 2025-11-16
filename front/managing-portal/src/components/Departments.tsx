import { useEffect, useState } from 'react';
import { departmentsApi } from '../services/departments';
import type { Department, DepartmentTreeNode, DepartmentWithStats } from '../services/departments';
import './Departments.css';
import { LuBuilding2, LuPlus, LuPencil, LuTrash2, LuUsers, LuChevronRight, LuChevronDown } from 'react-icons/lu';
import { SearchableSelect } from './SearchableSelect';

interface Organization {
  id: string;
  name: string;
  description?: string;
}

export const Departments = () => {
  const [departments, setDepartments] = useState<Department[]>([]);
  const [departmentTree, setDepartmentTree] = useState<DepartmentTreeNode | null>(null);
  const [viewMode, setViewMode] = useState<'list' | 'tree'>('tree');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [editingDepartment, setEditingDepartment] = useState<Department | null>(null);
  const [selectedDepartment, setSelectedDepartment] = useState<DepartmentWithStats | null>(null);
  const [expandedNodes, setExpandedNodes] = useState<Set<string>>(new Set());
  const [organizations, setOrganizations] = useState<Organization[]>([]);

  // Form state
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    parent_id: '',
    organization_id: '',
  });

  useEffect(() => {
    loadDepartments();
    loadOrganizations();
  }, [viewMode]);

  const loadOrganizations = async () => {
    try {
      const token = localStorage.getItem('token') || sessionStorage.getItem('token');
      const response = await fetch('/api/v1/organizations', {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (response.ok) {
        const data = await response.json();
        setOrganizations(data || []);
      }
    } catch (err) {
      console.error('Failed to fetch organizations:', err);
    }
  };

  const loadDepartments = async () => {
    setLoading(true);
    setError(null);
    try {
      // Always load departments list for the parent selector
      const depts = await departmentsApi.getDepartments();
      setDepartments(depts);

      // Additionally load tree if in tree view mode
      if (viewMode === 'tree') {
        const tree = await departmentsApi.getDepartmentTree();
        setDepartmentTree(tree);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load departments');
    } finally {
      setLoading(false);
    }
  };

  const handleCreateDepartment = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await departmentsApi.createDepartment({
        name: formData.name,
        description: formData.description,
        parent_id: formData.parent_id || undefined,
        organization_id: formData.organization_id || undefined,
      });
      setShowCreateForm(false);
      setFormData({ name: '', description: '', parent_id: '', organization_id: '' });
      loadDepartments();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to create department');
    }
  };

  const handleUpdateDepartment = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editingDepartment) return;

    try {
      await departmentsApi.updateDepartment(editingDepartment.id, {
        name: formData.name,
        description: formData.description,
        parent_id: formData.parent_id || undefined,
        organization_id: formData.organization_id || undefined,
      });
      setShowCreateForm(false);
      setEditingDepartment(null);
      setFormData({ name: '', description: '', parent_id: '', organization_id: '' });
      loadDepartments();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to update department');
    }
  };

  const handleDeleteDepartment = async (id: string, name: string) => {
    if (!confirm(`Are you sure you want to delete department "${name}"?`)) return;

    try {
      await departmentsApi.deleteDepartment(id);
      loadDepartments();
      if (selectedDepartment?.id === id) {
        setSelectedDepartment(null);
      }
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to delete department');
    }
  };

  const handleSelectDepartment = async (id: string) => {
    try {
      const dept = await departmentsApi.getDepartment(id, true) as DepartmentWithStats;
      setSelectedDepartment(dept);
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to load department details');
    }
  };

  const handleEditDepartment = (dept: Department) => {
    setEditingDepartment(dept);
    setFormData({
      name: dept.name,
      description: dept.description,
      parent_id: dept.parent_id || '',
      organization_id: (dept as any).organization_id || '',
    });
    setShowCreateForm(true);
  };

  const toggleNode = (nodeId: string) => {
    const newExpanded = new Set(expandedNodes);
    if (newExpanded.has(nodeId)) {
      newExpanded.delete(nodeId);
    } else {
      newExpanded.add(nodeId);
    }
    setExpandedNodes(newExpanded);
  };

  const renderTreeNode = (node: DepartmentTreeNode, depth: number = 0) => {
    const hasChildren = node.children && node.children.length > 0;
    const isExpanded = expandedNodes.has(node.id);
    const isSelected = selectedDepartment?.id === node.id;

    return (
      <div key={node.id} className="tree-node" style={{ marginLeft: `${depth * 20}px` }}>
        <div className={`tree-node-content ${isSelected ? 'selected' : ''}`}>
          <button
            className="expand-button"
            onClick={() => toggleNode(node.id)}
            disabled={!hasChildren}
          >
            {hasChildren ? (
              isExpanded ? <LuChevronDown /> : <LuChevronRight />
            ) : (
              <span style={{ width: '16px', display: 'inline-block' }}></span>
            )}
          </button>

          <div className="tree-node-info" onClick={() => handleSelectDepartment(node.id)}>
            <LuBuilding2 className="department-icon" />
            <span className="department-name">{node.name}</span>
            <span className="department-level">Level {node.level}</span>
          </div>

          <div className="tree-node-actions">
            <button
              className="icon-button edit"
              onClick={() => handleEditDepartment(node)}
              title="Edit department"
            >
              <LuPencil />
            </button>
            <button
              className="icon-button delete"
              onClick={() => handleDeleteDepartment(node.id, node.name)}
              title="Delete department"
            >
              <LuTrash2 />
            </button>
          </div>
        </div>

        {hasChildren && isExpanded && (
          <div className="tree-node-children">
            {node.children!.map(child => renderTreeNode(child, depth + 1))}
          </div>
        )}
      </div>
    );
  };

  if (loading) {
    return (
      <div className="departments-container">
        <div className="loading">Loading departments...</div>
      </div>
    );
  }

  return (
    <div className="departments-container">
      <div className="departments-header">
        <div className="header-title">
          <LuBuilding2 size={24} />
          <h1>Departments</h1>
        </div>

        <div className="header-actions">
          <div className="view-toggle">
            <button
              className={viewMode === 'list' ? 'active' : ''}
              onClick={() => setViewMode('list')}
            >
              List View
            </button>
            <button
              className={viewMode === 'tree' ? 'active' : ''}
              onClick={() => setViewMode('tree')}
            >
              Tree View
            </button>
          </div>

          <button
            className="create-button"
            onClick={() => {
              setShowCreateForm(true);
              setEditingDepartment(null);
              setFormData({ name: '', description: '', parent_id: '', organization_id: '' });
            }}
          >
            <LuPlus /> Create Department
          </button>
        </div>
      </div>

      {error && (
        <div className="error-message">
          {error}
          <button onClick={loadDepartments}>Retry</button>
        </div>
      )}

      <div className="departments-content">
        <div className="departments-main">
          {viewMode === 'tree' && departmentTree ? (
            <div className="tree-view">
              {renderTreeNode(departmentTree)}
            </div>
          ) : (
            <div className="list-view">
              {departments.length === 0 ? (
                <div className="empty-state">
                  <LuBuilding2 size={48} />
                  <p>No departments found</p>
                </div>
              ) : (
                <div className="department-cards">
                  {departments.map(dept => (
                    <div
                      key={dept.id}
                      className={`department-card ${selectedDepartment?.id === dept.id ? 'selected' : ''}`}
                      onClick={() => handleSelectDepartment(dept.id)}
                    >
                      <div className="card-header">
                        <LuBuilding2 />
                        <h3>{dept.name}</h3>
                        <span className="level-badge">Level {dept.level}</span>
                      </div>

                      {dept.description && (
                        <p className="card-description">{dept.description}</p>
                      )}

                      <div className="card-meta">
                        <span className="path">{dept.path}</span>
                      </div>

                      <div className="card-actions">
                        <button
                          className="icon-button edit"
                          onClick={(e) => {
                            e.stopPropagation();
                            handleEditDepartment(dept);
                          }}
                        >
                          <LuPencil /> Edit
                        </button>
                        <button
                          className="icon-button delete"
                          onClick={(e) => {
                            e.stopPropagation();
                            handleDeleteDepartment(dept.id, dept.name);
                          }}
                        >
                          <LuTrash2 /> Delete
                        </button>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}
        </div>

        {selectedDepartment && (
          <div className="department-details">
            <div className="details-header">
              <h2>{selectedDepartment.name}</h2>
              <button
                className="close-button"
                onClick={() => setSelectedDepartment(null)}
              >
                ×
              </button>
            </div>

            <div className="details-content">
              <div className="detail-section">
                <label>Description</label>
                <p>{selectedDepartment.description || 'No description'}</p>
              </div>

              <div className="detail-section">
                <label>Hierarchy</label>
                <p><strong>Level:</strong> {selectedDepartment.level}</p>
                <p><strong>Path:</strong> {selectedDepartment.path}</p>
              </div>

              <div className="detail-section">
                <label>Statistics</label>
                <div className="stats-grid">
                  <div className="stat-item">
                    <LuUsers />
                    <div>
                      <span className="stat-value">{selectedDepartment.user_count}</span>
                      <span className="stat-label">Direct Users</span>
                    </div>
                  </div>
                  <div className="stat-item">
                    <LuBuilding2 />
                    <div>
                      <span className="stat-value">{selectedDepartment.child_count}</span>
                      <span className="stat-label">Child Departments</span>
                    </div>
                  </div>
                  <div className="stat-item">
                    <LuUsers />
                    <div>
                      <span className="stat-value">{selectedDepartment.total_users_count}</span>
                      <span className="stat-label">Total Users (incl. sub-depts)</span>
                    </div>
                  </div>
                </div>
              </div>

              <div className="detail-section">
                <label>Status</label>
                <span className={`status-badge ${selectedDepartment.is_active ? 'active' : 'inactive'}`}>
                  {selectedDepartment.is_active ? 'Active' : 'Inactive'}
                </span>
              </div>
            </div>
          </div>
        )}
      </div>

      {showCreateForm && (
        <div className="modal-overlay" onClick={() => setShowCreateForm(false)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h2>{editingDepartment ? 'Edit Department' : 'Create Department'}</h2>
              <button className="close-button" onClick={() => setShowCreateForm(false)}>×</button>
            </div>

            <form onSubmit={editingDepartment ? handleUpdateDepartment : handleCreateDepartment}>
              <div className="form-group">
                <label htmlFor="name">Name *</label>
                <input
                  id="name"
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  required
                  placeholder="Enter department name"
                />
              </div>

              <div className="form-group">
                <label htmlFor="description">Description</label>
                <textarea
                  id="description"
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                  placeholder="Enter department description"
                  rows={3}
                />
              </div>

              <div className="form-group">
                <label htmlFor="organization">Organization *</label>
                <select
                  id="organization"
                  value={formData.organization_id}
                  onChange={(e) => setFormData({ ...formData, organization_id: e.target.value })}
                  required
                >
                  <option value="">Select an organization...</option>
                  {organizations.map((org) => (
                    <option key={org.id} value={org.id}>
                      {org.name}
                    </option>
                  ))}
                </select>
              </div>

              <div className="form-group">
                <SearchableSelect
                  id="parent_id"
                  label="Parent Department"
                  value={formData.parent_id}
                  onChange={(value) => setFormData({ ...formData, parent_id: value })}
                  options={departments
                    .filter(d => !editingDepartment || d.id !== editingDepartment.id)
                    .map(dept => ({
                      value: dept.id,
                      label: dept.path || dept.name,
                    }))}
                  placeholder="Select parent department..."
                  emptyPlaceholder="No parent (root level)"
                />
              </div>

              <div className="form-actions">
                <button type="button" onClick={() => setShowCreateForm(false)}>
                  Cancel
                </button>
                <button type="submit" className="primary">
                  {editingDepartment ? 'Update' : 'Create'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

export default Departments;
