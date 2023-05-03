<template>
  <div :class="{ 'has-logo': showLogo }"
    :style="{ backgroundColor: settings.sideTheme === 'theme-dark' ? variables.menuBg : variables.menuLightBg }">
    <logo v-if="showLogo" :collapse="false" />

    <el-scrollbar style="padding-left: 30px" :class="settings.sideTheme" wrap-class="scrollbar-wrapper">
      <el-menu :default-active="activeMenu" :collapse="false"
        :background-color="settings.sideTheme === 'theme-dark' ? variables.menuBg : variables.menuLightBg"
        :text-color="settings.sideTheme === 'theme-dark' ? variables.menuText : 'rgba(0,0,0,.65)'" :unique-opened="true"
        :active-text-color="settings.theme" :collapse-transition="false" mode="vertical">
        <sidebar-item v-for="(route, index) in sidebarRouters" :key="route.path + index" :item="route"
          :base-path="route.path" />
      </el-menu>
    </el-scrollbar>
  </div>
</template>

<script>
import { mapGetters, mapState } from "vuex";
import Logo from "./Logo";
import SidebarItem from "./SidebarItem";
import variables from "@/assets/styles/variables.scss";

export default {
  name: "sidebar",
  data() {
    return {

    }
  },
  components: { SidebarItem, Logo },
  computed: {
    ...mapState(["settings"]),
    ...mapGetters(["sidebarRouters", "sidebar"]),
    activeMenu() {
      const route = this.$route;
      const { meta, path } = route;
      // if set path, the sidebar will highlight the path you set
      if (meta.activeMenu) {
        return meta.activeMenu;
      }
      return path;
    },
    showLogo() {
      return this.$store.state.settings.sidebarLogo;
    },
    variables() {
      return variables;
    },
    isCollapse() {
      return !this.sidebar.opened;
    }
  },
};
</script>
<style lang="scss">
.cus-dropdown-menu {
  line-height: 38px;
  margin-left: 40px;

  .el-dropdown-menu__item {
    font-size: 18px !important;
    line-height: 38px;
    padding: 0 25px;
  }
}

.cus-dropdown {
  display: inline-block;
  height: 56px;
  line-height: 56px;
  margin-left: 50px;
  vertical-align: top;

  .el-dropdown {
    color: #97a8be;
    font-size: 18px !important;
  }
}
</style>
