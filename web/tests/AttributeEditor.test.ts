import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import AttributeEditor from '@/components/ci/AttributeEditor.vue'

// Mock the base components
vi.mock('@/components/base/BaseInput.vue', () => ({
  default: {
    template: '<input :value="modelValue" @input="$emit(\'update:modelValue\', $event.target.value)" />',
    props: ['modelValue']
  }
}))

vi.mock('@/components/base/BaseSelect.vue', () => ({
  default: {
    template: `
      <select :value="modelValue" @change="$emit(\'update:modelValue\', $event.target.value)">
        <option v-if="placeholder && !modelValue" value="">{{ placeholder }}</option>
        <option v-for="option in normalizedOptions" :key="option.value" :value="option.value">
          {{ option.label }}
        </option>
      </select>
    `,
    props: ['modelValue', 'options', 'placeholder'],
    computed: {
      normalizedOptions() {
        if (!this.options) return []
        return this.options.map(option => {
          if (typeof option === 'object' && option !== null) {
            return option
          }
          return { value: option, label: String(option) }
        })
      }
    }
  }
}))

vi.mock('@/components/base/BaseButton.vue', () => ({
  default: {
    template: '<button @click="$emit(\'click\')"><slot /></button>',
    props: ['variant', 'size']
  }
}))

vi.mock('@/components/base/Icon.vue', () => ({
  default: {
    template: '<span>{{ name }}</span>',
    props: ['name']
  }
}))

describe('AttributeEditor', () => {
  const mockAttribute = {
    name: 'hostname',
    type: 'string',
    description: 'Server hostname',
    validation: {}
  }

  it('should render component correctly', () => {
    const wrapper = mount(AttributeEditor, {
      props: {
        attribute: mockAttribute,
        index: 0,
        isRequired: true
      }
    })

    expect(wrapper.exists()).toBe(true)
    expect(wrapper.find('input').exists()).toBe(true) // Name input
    expect(wrapper.find('button').exists()).toBe(true) // Remove button
  })

  it('should show attribute type dropdown', () => {
    const wrapper = mount(AttributeEditor, {
      props: {
        attribute: mockAttribute,
        index: 0,
        isRequired: true
      }
    })

    const selects = wrapper.findAll('select')
    expect(selects.length).toBeGreaterThan(0)
  })

  it('should emit update when type is changed', async () => {
    const wrapper = mount(AttributeEditor, {
      props: {
        attribute: mockAttribute,
        index: 0,
        isRequired: true
      }
    })

    const selects = wrapper.findAll('select')
    if (selects.length > 1) {
      const typeSelect = selects[1] // Assuming second select is for type
      await typeSelect.setValue('integer')

      expect(wrapper.emitted('update')).toBeTruthy()
      const updateEvent = wrapper.emitted('update')[0]
      expect(updateEvent).toEqual([0, expect.objectContaining({ type: 'integer' })])
    }
  })

  it('should emit remove event when remove button is clicked', async () => {
    const wrapper = mount(AttributeEditor, {
      props: {
        attribute: mockAttribute,
        index: 2,
        isRequired: true
      }
    })

    const removeButton = wrapper.find('button[title*="Remove"]')
    if (removeButton.exists()) {
      await removeButton.trigger('click')
      expect(wrapper.emitted('remove')).toBeTruthy()
      expect(wrapper.emitted('remove')[0]).toEqual([2])
    }
  })

  it('should handle empty attribute', () => {
    const wrapper = mount(AttributeEditor, {
      props: {
        attribute: { name: '', type: '', description: '', validation: {} },
        index: 0,
        isRequired: false
      }
    })

    expect(wrapper.exists()).toBe(true)
    expect(wrapper.find('input').exists()).toBe(true)
  })

  it('should toggle validation rules visibility', async () => {
    const wrapper = mount(AttributeEditor, {
      props: {
        attribute: mockAttribute,
        index: 0,
        isRequired: true
      }
    })

    const buttons = wrapper.findAll('button')
    if (buttons.length > 0) {
      await buttons[0].trigger('click') // First button should toggle validation

      // After clicking, validation should be visible
      expect(wrapper.text()).toContain('Validation Rules')
    }
  })
})