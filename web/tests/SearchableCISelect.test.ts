import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'

const { listMock } = vi.hoisted(() => ({ listMock: vi.fn() }))

vi.mock('@/services/api', () => ({
  ciAPI: { list: listMock }
}))

import SearchableCISelect from '@/components/ci/SearchableCISelect.vue'

const makeResponse = (
  cis: Array<{ id: string; name: string; ci_type: string }>,
  page = 1,
  totalPages = 1
) => ({ data: { cis, page, limit: 100, total: cis.length, total_pages: totalPages } })

describe('SearchableCISelect', () => {
  beforeEach(() => {
    listMock.mockReset()
  })

  it('loads configuration items from the API and renders them as options', async () => {
    listMock.mockResolvedValue(
      makeResponse([
        { id: 'ci-1', name: 'server01', ci_type: 'Server' },
        { id: 'ci-2', name: 'db01', ci_type: 'Database' }
      ])
    )

    const wrapper = mount(SearchableCISelect, { props: { modelValue: null } })
    await flushPromises()

    expect(listMock).toHaveBeenCalledWith({ page: 1, limit: 100 })

    const options = wrapper.findAll('option')
    // Placeholder for the single-select plus one option per CI.
    expect(options).toHaveLength(3)
    expect(options.map((o) => o.text())).toEqual(
      expect.arrayContaining(['server01 (Server)', 'db01 (Database)'])
    )
  })

  it('groups options by CI type', async () => {
    listMock.mockResolvedValue(
      makeResponse([
        { id: 'ci-1', name: 'server01', ci_type: 'Server' },
        { id: 'ci-2', name: 'server02', ci_type: 'Server' },
        { id: 'ci-3', name: 'db01', ci_type: 'Database' }
      ])
    )

    const wrapper = mount(SearchableCISelect, { props: { modelValue: null } })
    await flushPromises()

    const groups = wrapper.findAll('optgroup')
    expect(groups.map((g) => g.attributes('label'))).toEqual(['Database', 'Server'])
  })

  it('loads every page of results', async () => {
    listMock
      .mockResolvedValueOnce(
        makeResponse([{ id: 'ci-1', name: 'server01', ci_type: 'Server' }], 1, 2)
      )
      .mockResolvedValueOnce(
        makeResponse([{ id: 'ci-2', name: 'db01', ci_type: 'Database' }], 2, 2)
      )

    const wrapper = mount(SearchableCISelect, { props: { modelValue: null } })
    await flushPromises()

    expect(listMock).toHaveBeenCalledTimes(2)
    const options = wrapper.findAll('option').map((o) => o.text())
    expect(options).toEqual(expect.arrayContaining(['server01 (Server)', 'db01 (Database)']))
  })

  it('marks excluded configuration items as disabled', async () => {
    listMock.mockResolvedValue(
      makeResponse([
        { id: 'ci-1', name: 'server01', ci_type: 'Server' },
        { id: 'ci-2', name: 'db01', ci_type: 'Database' }
      ])
    )

    const wrapper = mount(SearchableCISelect, {
      props: { modelValue: null, excludeIds: ['ci-1'] }
    })
    await flushPromises()

    const excluded = wrapper.find('option[value="ci-1"]')
    const allowed = wrapper.find('option[value="ci-2"]')
    expect(excluded.attributes('disabled')).toBeDefined()
    expect(allowed.attributes('disabled')).toBeUndefined()
  })

  it('surfaces a load failure instead of silently rendering an empty list', async () => {
    vi.spyOn(console, 'error').mockImplementation(() => {})
    listMock.mockRejectedValue(new Error('boom'))

    const wrapper = mount(SearchableCISelect, { props: { modelValue: null } })
    await flushPromises()

    expect(wrapper.text()).toContain('Failed to load configuration items')
  })
})
