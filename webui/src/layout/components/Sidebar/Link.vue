<template>
  <component :is="type" v-bind="linkProps(to)">
    <slot />
  </component>
</template>

<script>
import { isExternal } from '@/utils/validate'

export default {
  props: {
    to: {
      type: Object,
      required: true
    }
  },
  computed: {
    isExternal() {
      return isExternal(this.to.path)
    },
    type() {
      if (this.isExternal) {
        return 'a'
      }
      return 'router-link'
    }
  },
  methods: {
    linkProps(to) {
      if (this.isExternal) {
        return {
          href: to.path,
          target: '_blank',
          rel: 'noopener'
        }
      }
      if (to.hasParams){
        return {
          to: {
            path: to.path,
            query: to.params
          }
        }
      } else {
        return {
          to: {
            name: to.name
          }
        }
      }
    }
  }
}
</script>
