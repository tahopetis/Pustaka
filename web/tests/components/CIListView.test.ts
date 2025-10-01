import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import CIListView from '@/views/ci/CIListView.vue'
import { useCIStore } from '@/stores/ci'
import { useAuthStore } from '@/stores/auth'

// Mock the router
vi.mock('vue-router', () => ({
  useRouter: () => ({
    push: vi.fn()
  })
}))

// Mock the store
vi.mock('@/stores/ci')
vi.mock('@/stores/auth')

describe('CIListView', () => {
  let wrapper: any
  let mockCIStore: any
  let mockAuthStore: any

  beforeEach(() => {
    setActivePinia(createPinia())

    // Mock CI store
    mockCIStore = {
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
      ciTypes: [
        { id: '1', name: 'Server' },
        { id: '2', name: 'Database' },
        { id: '3', name: 'Application' }
      ],
      loading: false,
      error: null,
      pagination: {
        page: 1,
        limit: 20,
        total: 2,
        totalPages: 1
      },
      filters: {
        ciType: '',
        search: '',
        tags: [],
        sortBy: 'name',
        sortOrder: 'asc'
      },
      fetchCIs: vi.fn(),
      fetchCITypes: vi.fn(),
      deleteCI: vi.fn(),
      setFilters: vi.fn(),
      setPagination: vi.fn(),
      resetFilters: vi.fn()
    }

    // Mock Auth store
    mockAuthStore = {
      isAuthenticated: true,
      hasPermission: vi.fn().mockReturnValue(true)
    }

    vi.mocked(useCIStore).mockReturnValue(mockCIStore)
    vi.mocked(useAuthStore).mockReturnValue(mockAuthStore)

    wrapper = mount(CIListView)
  })

  it('renders CI list correctly', () => {
    expect(wrapper.find('h1').text()).toBe('Configuration Items')
    expect(wrapper.find('[data-test="search-input"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="ci-type-filter"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="create-ci-button"]').exists()).toBe(true)
  })

  it('displays CI items in table', () => {
    const rows = wrapper.findAll('tbody tr')
    expect(rows).toHaveLength(2)

    const firstRow = rows[0]
    expect(firstRow.text()).toContain('web-server-01')
    expect(firstRow.text()).toContain('Server')
    expect(firstRow.text()).toContain('production')
    expect(firstRow.text()).toContain('web')

    const secondRow = rows[1]
    expect(secondRow.text()).toContain('database-01')
    expect(secondRow.text()).toContain('Database')
  })

  it('shows loading state', async () => {
    mockCIStore.loading = true
    await wrapper.vm.$nextTick()

    expect(wrapper.find('[data-test="loading-spinner"]').exists()).toBe(true)
    expect(wrapper.find('table').exists()).toBe(false)
  })

  it('shows error message', async () => {
    mockCIStore.error = 'Failed to fetch CIs'
    await wrapper.vm.$nextTick()

    expect(wrapper.find('[data-test="error-message"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="error-message"]').text()).toContain('Failed to fetch CIs')
  })

  it('shows empty state when no CIs', async () => {
    mockCIStore.cis = []
    await wrapper.vm.$nextTick()

    expect(wrapper.find('[data-test="empty-state"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="empty-state"]').text()).toContain('No configuration items found')
  })

  it('populates CI type filter', () => {
    const selectElement = wrapper.find('[data-test="ci-type-filter"]')
    const options = selectElement.findAll('option')

    // Should have "All Types" option + CI types
    expect(options).toHaveLength(4) // All Types + 3 CI types
    expect(options[0].text()).toBe('All Types')
    expect(options[1].text()).toBe('Server')
    expect(options[2].text()).toBe('Database')
    expect(options[3].text()).toBe('Application')
  })

  it('handles search input', async () => {
    const searchInput = wrapper.find('[data-test="search-input"]')
    await searchInput.setValue('web-server')
    await searchInput.trigger('input')

    // Debounced search - wait for it
    await new Promise(resolve => setTimeout(resolve, 350))

    expect(mockCIStore.setFilters).toHaveBeenCalledWith(
      expect.objectContaining({ search: 'web-server' })
    )
  })

  it('handles CI type filter change', async () => {
    const selectElement = wrapper.find('[data-test="ci-type-filter"]')
    await selectElement.setValue('Server')
    await selectElement.trigger('change')

    expect(mockCIStore.setFilters).toHaveBeenCalledWith(
      expect.objectContaining({ ciType: 'Server' })
    )
  })

  it('handles sort change', async () => {
    const sortSelect = wrapper.find('[data-test="sort-select"]')
    await sortSelect.setValue('created_at')
    await sortSelect.trigger('change')

    expect(mockCIStore.setFilters).toHaveBeenCalledWith(
      expect.objectContaining({ sortBy: 'created_at' })
    )
  })

  it('handles page change', async () => {
    const pagination = wrapper.find('[data-test="pagination"]')
    const nextPageButton = pagination.find('[data-test="next-page"]')
    await nextPageButton.trigger('click')

    expect(mockCIStore.setPagination).toHaveBeenCalledWith(
      expect.objectContaining({ page: 2 })
    )
  })

  it('shows edit and delete buttons for authenticated users with permissions', () => {
    const actionButtons = wrapper.findAll('[data-test="action-buttons"]')
    expect(actionButtons).toHaveLength(2)

    const firstRowActions = actionButtons[0]
    expect(firstRowActions.find('[data-test="edit-button"]').exists()).toBe(true)
    expect(firstRowActions.find('[data-test="delete-button"]').exists()).toBe(true)
  })

  it('hides action buttons for users without permissions', async () => {
    mockAuthStore.hasPermission.mockReturnValue(false)
    await wrapper.vm.$nextTick()

    const actionButtons = wrapper.findAll('[data-test="action-buttons"]')
    expect(actionButtons).toHaveLength(0)
  })

  it('shows create button for users with create permission', () => {
    const createButton = wrapper.find('[data-test="create-ci-button"]')
    expect(createButton.exists()).toBe(true)
    expect(createButton.text()).toContain('Create CI')
  })

  it('hides create button for users without create permission', async () => {
    mockAuthStore.hasPermission.mockImplementation((permission: string) => {
      return permission !== 'ci:create'
    })
    await wrapper.vm.$nextTick()

    const createButton = wrapper.find('[data-test="create-ci-button"]')
    expect(createButton.exists()).toBe(false)
  })

  it('confirms CI deletion', async () => {
    const deleteButton = wrapper.find('[data-test="delete-button"]')
    await deleteButton.trigger('click')

    // Mock window.confirm
    const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(true)

    await wrapper.vm.$nextTick()

    expect(confirmSpy).toHaveBeenCalledWith(
      'Are you sure you want to delete "web-server-01"?'
    )

    expect(mockCIStore.deleteCI).toHaveBeenCalledWith('550e8400-e29b-41d4-a716-446655440000')

    confirmSpy.mockRestore()
  })

  it('cancels CI deletion', async () => {
    const deleteButton = wrapper.find('[data-test="delete-button"]')
    await deleteButton.trigger('click')

    // Mock window.confirm to return false
    const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(false)

    await wrapper.vm.$nextTick()

    expect(confirmSpy).toHaveBeenCalled()
    expect(mockCIStore.deleteCI).not.toHaveBeenCalled()

    confirmSpy.mockRestore()
  })

  it('handles tag display correctly', () => {
    const tags = wrapper.findAll('[data-test="ci-tags"]')
    expect(tags).toHaveLength(2)

    const firstRowTags = tags[0].findAll('[data-test="tag"]')
    expect(firstRowTags).toHaveLength(2)
    expect(firstRowTags[0].text()).toBe('production')
    expect(firstRowTags[1].text()).toBe('web')

    const secondRowTags = tags[1].findAll('[data-test="tag"]')
    expect(secondRowTags).toHaveLength(2)
    expect(secondRowTags[0].text()).toBe('production')
    expect(secondRowTags[1].text()).toBe('database')
  })

  it('formats dates correctly', () => {
    const dates = wrapper.findAll('[data-test="created-date"]')
    expect(dates).toHaveLength(2)
    expect(dates[0].text()).toBe('Jan 1, 2023')
    expect(dates[1].text()).toBe('Jan 1, 2023')
  })

  it('triggers data refresh on mount', () => {
    expect(mockCIStore.fetchCIs).toHaveBeenCalled()
    expect(mockCIStore.fetchCITypes).toHaveBeenCalled()
  })

  it('handles reset filters', async () => {
    const resetButton = wrapper.find('[data-test="reset-filters"]')
    await resetButton.trigger('click')

    expect(mockCIStore.resetFilters).toHaveBeenCalled()
  })

  it('displays pagination information correctly', () => {
    const paginationInfo = wrapper.find('[data-test="pagination-info"]')
    expect(paginationInfo.text()).toContain('Showing 1-2 of 2 items')
  })
})