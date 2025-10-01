import { describe, it, expect, beforeEach, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useCIStore } from '@/stores/ci'
import axios from 'axios'

// Mock axios
vi.mock('axios')

describe('CI Store', () => {
  let ciStore: any

  beforeEach(() => {
    setActivePinia(createPinia())
    ciStore = useCIStore()
    vi.clearAllMocks()
  })

  describe('Initial State', () => {
    it('should have correct initial state', () => {
      expect(ciStore.cis).toEqual([])
      expect(ciStore.ciTypes).toEqual([])
      expect(ciStore.currentCI).toBeNull()
      expect(ciStore.relationships).toEqual([])
      expect(ciStore.loading).toBe(false)
      expect(ciStore.error).toBeNull()
      expect(ciStore.pagination).toEqual({
        page: 1,
        limit: 20,
        total: 0,
        totalPages: 0
      })
      expect(ciStore.filters).toEqual({
        ciType: '',
        search: '',
        tags: [],
        sortBy: 'name',
        sortOrder: 'asc'
      })
    })
  })

  describe('Fetch CIs', () => {
    it('should fetch CIs successfully', async () => {
      const mockResponse = {
        data: {
          cis: [
            {
              id: '550e8400-e29b-41d4-a716-446655440000',
              name: 'web-server-01',
              ci_type: 'Server',
              attributes: {
                hostname: 'web-server-01',
                ip_address: '192.168.1.10',
                os: 'Ubuntu 20.04'
              },
              tags: ['production', 'web'],
              created_at: '2023-01-01T00:00:00Z',
              updated_at: '2023-01-01T00:00:00Z'
            },
            {
              id: '550e8400-e29b-41d4-a716-446655440001',
              name: 'database-01',
              ci_type: 'Database',
              attributes: {
                hostname: 'database-01',
                engine: 'PostgreSQL',
                version: '14'
              },
              tags: ['production', 'database'],
              created_at: '2023-01-01T00:00:00Z',
              updated_at: '2023-01-01T00:00:00Z'
            }
          ],
          total: 2,
          page: 1,
          limit: 20,
          total_pages: 1
        }
      }

      vi.mocked(axios.get).mockResolvedValueOnce(mockResponse)

      await ciStore.fetchCIs()

      expect(ciStore.loading).toBe(false)
      expect(ciStore.cis).toHaveLength(2)
      expect(ciStore.cis[0].name).toBe('web-server-01')
      expect(ciStore.cis[1].ci_type).toBe('Database')
      expect(ciStore.pagination.total).toBe(2)
      expect(ciStore.error).toBeNull()
    })

    it('should handle fetch CIs error', async () => {
      const mockError = new Error('Network error')
      vi.mocked(axios.get).mockRejectedValueOnce(mockError)

      await ciStore.fetchCIs()

      expect(ciStore.loading).toBe(false)
      expect(ciStore.cis).toEqual([])
      expect(ciStore.error).toBe('Network error')
    })

    it('should apply filters when fetching CIs', async () => {
      const mockResponse = {
        data: {
          cis: [
            {
              id: '550e8400-e29b-41d4-a716-446655440000',
              name: 'web-server-01',
              ci_type: 'Server',
              attributes: { hostname: 'web-server-01' },
              tags: ['production'],
              created_at: '2023-01-01T00:00:00Z',
              updated_at: '2023-01-01T00:00:00Z'
            }
          ],
          total: 1,
          page: 1,
          limit: 20,
          total_pages: 1
        }
      }

      vi.mocked(axios.get).mockResolvedValueOnce(mockResponse)

      // Set filters
      ciStore.setFilters({
        ciType: 'Server',
        search: 'web',
        tags: ['production']
      })

      await ciStore.fetchCIs()

      expect(vi.mocked(axios.get)).toHaveBeenCalledWith(
        expect.stringContaining('ci_type=Server'),
        expect.any(Object)
      )
      expect(ciStore.cis).toHaveLength(1)
      expect(ciStore.cis[0].ci_type).toBe('Server')
    })
  })

  describe('Create CI', () => {
    it('should create CI successfully', async () => {
      const mockCI = {
        id: '550e8400-e29b-41d4-a716-446655440000',
        name: 'new-server',
        ci_type: 'Server',
        attributes: {
          hostname: 'new-server-01',
          ip_address: '192.168.1.50'
        },
        tags: ['test'],
        created_at: '2023-01-01T00:00:00Z',
        updated_at: '2023-01-01T00:00:00Z'
      }

      const mockResponse = {
        data: mockCI
      }

      vi.mocked(axios.post).mockResolvedValueOnce(mockResponse)

      const ciData = {
        name: 'new-server',
        ci_type: 'Server',
        attributes: {
          hostname: 'new-server-01',
          ip_address: '192.168.1.50'
        },
        tags: ['test']
      }

      const result = await ciStore.createCI(ciData)

      expect(result).toEqual(mockCI)
      expect(ciStore.error).toBeNull()
    })

    it('should handle create CI validation error', async () => {
      const mockError = {
        response: {
          status: 400,
          data: { error: 'Name is required' }
        }
      }

      vi.mocked(axios.post).mockRejectedValueOnce(mockError)

      const ciData = {
        name: '',
        ci_type: 'Server',
        attributes: {},
        tags: []
      }

      await expect(ciStore.createCI(ciData)).rejects.toThrow()

      expect(ciStore.error).toBe('Name is required')
    })
  })

  describe('Update CI', () => {
    it('should update CI successfully', async () => {
      const mockCI = {
        id: '550e8400-e29b-41d4-a716-446655440000',
        name: 'updated-server',
        ci_type: 'Server',
        attributes: {
          hostname: 'updated-server-01',
          os: 'Ubuntu 22.04'
        },
        tags: ['production'],
        created_at: '2023-01-01T00:00:00Z',
        updated_at: '2023-01-02T00:00:00Z'
      }

      const mockResponse = {
        data: mockCI
      }

      vi.mocked(axios.put).mockResolvedValueOnce(mockResponse)

      const ciId = '550e8400-e29b-41d4-a716-446655440000'
      const updateData = {
        name: 'updated-server',
        attributes: {
          os: 'Ubuntu 22.04'
        }
      }

      const result = await ciStore.updateCI(ciId, updateData)

      expect(result).toEqual(mockCI)
      expect(vi.mocked(axios.put)).toHaveBeenCalledWith(
        `/ci/${ciId}`,
        updateData
      )
    })

    it('should handle update CI not found error', async () => {
      const mockError = {
        response: {
          status: 404,
          data: { error: 'Configuration item not found' }
        }
      }

      vi.mocked(axios.put).mockRejectedValueOnce(mockError)

      const ciId = '550e8400-e29b-41d4-a716-446655440999'
      const updateData = {
        name: 'updated-server'
      }

      await expect(ciStore.updateCI(ciId, updateData)).rejects.toThrow()

      expect(ciStore.error).toBe('Configuration item not found')
    })
  })

  describe('Delete CI', () => {
    it('should delete CI successfully', async () => {
      vi.mocked(axios.delete).mockResolvedValueOnce({})

      const ciId = '550e8400-e29b-41d4-a716-446655440000'

      await ciStore.deleteCI(ciId)

      expect(vi.mocked(axios.delete)).toHaveBeenCalledWith(`/ci/${ciId}`)
    })

    it('should handle delete CI with relationships error', async () => {
      const mockError = {
        response: {
          status: 409,
          data: { error: 'Cannot delete CI with existing relationships' }
        }
      }

      vi.mocked(axios.delete).mockRejectedValueOnce(mockError)

      const ciId = '550e8400-e29b-41d4-a716-446655440000'

      await expect(ciStore.deleteCI(ciId)).rejects.toThrow()

      expect(ciStore.error).toBe('Cannot delete CI with existing relationships')
    })
  })

  describe('CI Types', () => {
    it('should fetch CI types successfully', async () => {
      const mockResponse = {
        data: {
          ci_types: [
            {
              id: '550e8400-e29b-41d4-a716-446655440001',
              name: 'Server',
              description: 'Physical or virtual server',
              required_attributes: [
                { name: 'hostname', type: 'string', required: true }
              ],
              optional_attributes: [
                { name: 'ip_address', type: 'string', required: false }
              ],
              created_at: '2023-01-01T00:00:00Z'
            },
            {
              id: '550e8400-e29b-41d4-a716-446655440002',
              name: 'Application',
              description: 'Software application',
              required_attributes: [
                { name: 'name', type: 'string', required: true },
                { name: 'version', type: 'string', required: true }
              ],
              optional_attributes: [],
              created_at: '2023-01-01T00:00:00Z'
            }
          ]
        }
      }

      vi.mocked(axios.get).mockResolvedValueOnce(mockResponse)

      await ciStore.fetchCITypes()

      expect(ciStore.ciTypes).toHaveLength(2)
      expect(ciStore.ciTypes[0].name).toBe('Server')
      expect(ciStore.ciTypes[1].name).toBe('Application')
    })
  })

  describe('Relationships', () => {
    it('should fetch CI relationships successfully', async () => {
      const mockResponse = {
        data: [
          {
            id: '550e8400-e29b-41d4-a716-446655440010',
            source_id: '550e8400-e29b-41d4-a716-446655440000',
            target_id: '550e8400-e29b-41d4-a716-446655440001',
            relationship_type: 'depends_on',
            attributes: {
              description: 'Application depends on database'
            },
            created_at: '2023-01-01T00:00:00Z'
          }
        ]
      }

      vi.mocked(axios.get).mockResolvedValueOnce(mockResponse)

      const ciId = '550e8400-e29b-41d4-a716-446655440000'
      await ciStore.fetchCIRelationships(ciId)

      expect(ciStore.relationships).toHaveLength(1)
      expect(ciStore.relationships[0].relationship_type).toBe('depends_on')
      expect(vi.mocked(axios.get)).toHaveBeenCalledWith(
        `/ci/${ciId}/relationships`
      )
    })

    it('should create relationship successfully', async () => {
      const mockRelationship = {
        id: '550e8400-e29b-41d4-a716-446655440010',
        source_id: '550e8400-e29b-41d4-a716-446655440000',
        target_id: '550e8400-e29b-41d4-a716-446655440001',
        relationship_type: 'connects_to',
        attributes: {
          port: 5432
        },
        created_at: '2023-01-01T00:00:00Z'
      }

      const mockResponse = {
        data: mockRelationship
      }

      vi.mocked(axios.post).mockResolvedValueOnce(mockResponse)

      const relationshipData = {
        source_id: '550e8400-e29b-41d4-a716-446655440000',
        target_id: '550e8400-e29b-41d4-a716-446655440001',
        relationship_type: 'connects_to',
        attributes: {
          port: 5432
        }
      }

      const result = await ciStore.createRelationship(relationshipData)

      expect(result).toEqual(mockRelationship)
      expect(vi.mocked(axios.post)).toHaveBeenCalledWith(
        '/relationships',
        relationshipData
      )
    })
  })

  describe('Graph Data', () => {
    it('should fetch graph data successfully', async () => {
      const mockResponse = {
        data: {
          nodes: [
            {
              id: '550e8400-e29b-41d4-a716-446655440000',
              label: 'Web Server',
              type: 'Server',
              group: 'infrastructure',
              properties: {
                hostname: 'web-01',
                status: 'active'
              }
            },
            {
              id: '550e8400-e29b-41d4-a716-446655440001',
              label: 'Database',
              type: 'Database',
              group: 'infrastructure',
              properties: {
                engine: 'PostgreSQL',
                version: '14'
              }
            }
          ],
          edges: [
            {
              id: '550e8400-e29b-41d4-a716-446655440010',
              from: '550e8400-e29b-41d4-a716-446655440000',
              to: '550e8400-e29b-41d4-a716-446655440001',
              label: 'connects to',
              type: 'dependency',
              properties: {
                protocol: 'tcp',
                port: 5432
              }
            }
          ]
        }
      }

      vi.mocked(axios.get).mockResolvedValueOnce(mockResponse)

      const filters = {
        ci_type: 'Server',
        relationship: 'depends_on',
        depth: 2
      }

      const result = await ciStore.fetchGraphData(filters)

      expect(result.nodes).toHaveLength(2)
      expect(result.edges).toHaveLength(1)
      expect(vi.mocked(axios.get)).toHaveBeenCalledWith(
        '/graph',
        { params: filters }
      )
    })
  })

  describe('Filters and Pagination', () => {
    it('should update filters correctly', () => {
      const newFilters = {
        ciType: 'Server',
        search: 'web',
        tags: ['production', 'web'],
        sortBy: 'created_at',
        sortOrder: 'desc'
      }

      ciStore.setFilters(newFilters)

      expect(ciStore.filters).toEqual(newFilters)
    })

    it('should update pagination correctly', () => {
      const newPagination = {
        page: 2,
        limit: 50
      }

      ciStore.setPagination(newPagination)

      expect(ciStore.pagination.page).toBe(2)
      expect(ciStore.pagination.limit).toBe(50)
    })

    it('should reset filters', () => {
      ciStore.setFilters({
        ciType: 'Server',
        search: 'test',
        tags: ['test']
      })

      ciStore.resetFilters()

      expect(ciStore.filters.ciType).toBe('')
      expect(ciStore.filters.search).toBe('')
      expect(ciStore.filters.tags).toEqual([])
      expect(ciStore.filters.sortBy).toBe('name')
      expect(ciStore.filters.sortOrder).toBe('asc')
    })
  })
})