import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import BaseSelect from '@/components/base/BaseSelect.vue'

describe('BaseSelect', () => {
  describe('with object options', () => {
    it('should render options correctly with object array', () => {
      const options = [
        { value: 'string', label: 'String' },
        { value: 'integer', label: 'Integer' },
        { value: 'boolean', label: 'Boolean' },
        { value: 'array', label: 'Array' },
        { value: 'object', label: 'Object' }
      ]

      const wrapper = mount(BaseSelect, {
        props: {
          modelValue: '',
          options
        }
      })

      const selectOptions = wrapper.findAll('option')
      expect(selectOptions).toHaveLength(5) // 5 options

      // Test first option
      expect(selectOptions[0].attributes('value')).toBe('string')
      expect(selectOptions[0].text()).toBe('String')

      // Test last option
      expect(selectOptions[4].attributes('value')).toBe('object')
      expect(selectOptions[4].text()).toBe('Object')
    })

    it('should support placeholder with object options', () => {
      const options = [
        { value: 'string', label: 'String' },
        { value: 'integer', label: 'Integer' }
      ]

      const wrapper = mount(BaseSelect, {
        props: {
          modelValue: '',
          options,
          placeholder: 'Select type'
        }
      })

      const selectOptions = wrapper.findAll('option')
      expect(selectOptions).toHaveLength(3) // placeholder + 2 options

      expect(selectOptions[0].attributes('value')).toBe('')
      expect(selectOptions[0].text()).toBe('Select type')
    })

    it('should handle disabled options', () => {
      const options = [
        { value: 'string', label: 'String' },
        { value: 'integer', label: 'Integer', disabled: true },
        { value: 'boolean', label: 'Boolean' }
      ]

      const wrapper = mount(BaseSelect, {
        props: {
          modelValue: '',
          options
        }
      })

      const selectOptions = wrapper.findAll('option')
      expect(selectOptions[1].attributes('disabled')).toBe('')
      expect(selectOptions[0].attributes('disabled')).toBeUndefined()
      expect(selectOptions[2].attributes('disabled')).toBeUndefined()
    })
  })

  describe('with primitive options', () => {
    it('should render options correctly with string array', () => {
      const options = ['string', 'integer', 'boolean', 'array', 'object']

      const wrapper = mount(BaseSelect, {
        props: {
          modelValue: '',
          options
        }
      })

      const selectOptions = wrapper.findAll('option')
      expect(selectOptions).toHaveLength(5)

      expect(selectOptions[0].attributes('value')).toBe('string')
      expect(selectOptions[0].text()).toBe('string')

      expect(selectOptions[2].attributes('value')).toBe('boolean')
      expect(selectOptions[2].text()).toBe('boolean')
    })

    it('should render options correctly with number array', () => {
      const options = [1, 2, 3, 4, 5]

      const wrapper = mount(BaseSelect, {
        props: {
          modelValue: '',
          options
        }
      })

      const selectOptions = wrapper.findAll('option')
      expect(selectOptions).toHaveLength(5)

      expect(selectOptions[0].attributes('value')).toBe('1')
      expect(selectOptions[0].text()).toBe('1')
    })
  })

  describe('v-model binding', () => {
    it('should emit update:modelValue when option is selected', async () => {
      const options = [
        { value: 'string', label: 'String' },
        { value: 'integer', label: 'Integer' }
      ]

      const wrapper = mount(BaseSelect, {
        props: {
          modelValue: '',
          options
        }
      })

      const select = wrapper.find('select')
      await select.setValue('integer')

      expect(wrapper.emitted('update:modelValue')).toBeTruthy()
      expect(wrapper.emitted('update:modelValue')![0]).toEqual(['integer'])
    })

    it('should respect initial modelValue', () => {
      const options = [
        { value: 'string', label: 'String' },
        { value: 'integer', label: 'Integer' },
        { value: 'boolean', label: 'Boolean' }
      ]

      const wrapper = mount(BaseSelect, {
        props: {
          modelValue: 'integer',
          options
        }
      })

      const select = wrapper.find('select')
      expect(select.element.value).toBe('integer')
    })
  })

  describe('slot support', () => {
    it('should render slot content alongside options', () => {
      const options = [
        { value: 'string', label: 'String' },
        { value: 'integer', label: 'Integer' }
      ]

      const wrapper = mount(BaseSelect, {
        props: {
          modelValue: '',
          options
        },
        slots: {
          default: '<option value="custom">Custom Option</option>'
        }
      })

      const selectOptions = wrapper.findAll('option')
      expect(selectOptions).toHaveLength(3) // 2 from options + 1 from slot

      expect(selectOptions[2].attributes('value')).toBe('custom')
      expect(selectOptions[2].text()).toBe('Custom Option')
    })
  })

  describe('attributes', () => {
    it('should apply disabled state correctly', () => {
      const wrapper = mount(BaseSelect, {
        props: {
          modelValue: '',
          options: [{ value: 'test', label: 'Test' }],
          disabled: true
        }
      })

      const select = wrapper.find('select')
      expect(select.attributes('disabled')).toBe('')
    })

    it('should apply required state correctly', () => {
      const wrapper = mount(BaseSelect, {
        props: {
          modelValue: '',
          options: [{ value: 'test', label: 'Test' }],
          required: true
        }
      })

      const select = wrapper.find('select')
      expect(select.attributes('required')).toBe('')
    })

    it('should apply error state correctly', () => {
      const wrapper = mount(BaseSelect, {
        props: {
          modelValue: '',
          options: [{ value: 'test', label: 'Test' }],
          error: true
        }
      })

      const select = wrapper.find('select')
      expect(select.classes()).toContain('border-red-300')
      expect(select.classes()).toContain('text-red-900')
      expect(select.classes()).toContain('focus:ring-red-500')
      expect(select.classes()).toContain('focus:border-red-500')
    })

    it('should apply custom id correctly', () => {
      const wrapper = mount(BaseSelect, {
        props: {
          modelValue: '',
          options: [{ value: 'test', label: 'Test' }],
          id: 'test-select'
        }
      })

      const select = wrapper.find('select')
      expect(select.attributes('id')).toBe('test-select')
    })
  })

  describe('events', () => {
    it('should emit blur event', async () => {
      const wrapper = mount(BaseSelect, {
        props: {
          modelValue: '',
          options: [{ value: 'test', label: 'Test' }]
        }
      })

      const select = wrapper.find('select')
      await select.trigger('blur')

      expect(wrapper.emitted('blur')).toBeTruthy()
    })

    it('should emit focus event', async () => {
      const wrapper = mount(BaseSelect, {
        props: {
          modelValue: '',
          options: [{ value: 'test', label: 'Test' }]
        }
      })

      const select = wrapper.find('select')
      await select.trigger('focus')

      expect(wrapper.emitted('focus')).toBeTruthy()
    })
  })
})