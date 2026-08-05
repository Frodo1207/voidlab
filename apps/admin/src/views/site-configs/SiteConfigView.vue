<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import { ElMessage } from "element-plus";
import {
  defaultContactChannelsConfig,
  defaultFeaturedContentSlotsConfig,
  defaultFooterConfig,
  defaultGlobalCtaConfig,
  defaultHomeBannerConfig,
  defaultHomeFeaturedConfig,
  listSiteConfigs,
  updateSiteConfig
} from "../../services/siteConfigService";
import type {
  ContactChannelConfig,
  FeaturedContentSlotsConfig,
  FooterConfig,
  GlobalCtaConfig,
  HomeBannerConfig,
  HomeFeaturedConfig,
  SiteConfigKey,
  SiteConfigRecord
} from "../../types";

const loading = ref(false);
const savingKey = ref<SiteConfigKey | "">("");
const lastUpdatedMap = reactive<Record<string, string>>({});

const homeBannerForm = reactive<HomeBannerConfig>(cloneValue(defaultHomeBannerConfig));
const homeFeaturedForm = reactive<HomeFeaturedConfig>(cloneValue(defaultHomeFeaturedConfig));
const contactChannelsForm = ref<ContactChannelConfig[]>(cloneValue(defaultContactChannelsConfig));
const footerForm = reactive<FooterConfig>(cloneValue(defaultFooterConfig));
const globalCtaForm = reactive<GlobalCtaConfig>(cloneValue(defaultGlobalCtaConfig));
const featuredSlotsForm = reactive<FeaturedContentSlotsConfig>(cloneValue(defaultFeaturedContentSlotsConfig));

function cloneValue<T>(value: T): T {
  return JSON.parse(JSON.stringify(value)) as T;
}

function assignObject<T extends object>(target: T, source: T) {
  Object.assign(target, cloneValue(source));
}

function updateLastUpdated(records: SiteConfigRecord[]) {
  Object.keys(lastUpdatedMap).forEach((key) => {
    delete lastUpdatedMap[key];
  });

  records.forEach((record) => {
    lastUpdatedMap[record.configKey] = record.updatedAt;
  });
}

async function loadConfigs() {
  loading.value = true;
  try {
    const records = await listSiteConfigs();
    updateLastUpdated(records);

    const homeBanner = records.find((item) => item.configKey === "home_banner");
    const homeFeatured = records.find((item) => item.configKey === "home_featured");
    const contactChannels = records.find((item) => item.configKey === "contact_channels");
    const footerConfig = records.find((item) => item.configKey === "footer_config");
    const globalCta = records.find((item) => item.configKey === "global_cta");
    const featuredSlots = records.find((item) => item.configKey === "featured_content_slots");

    if (homeBanner) {
      assignObject(homeBannerForm, homeBanner.configValue as HomeBannerConfig);
    }

    if (homeFeatured) {
      assignObject(homeFeaturedForm, homeFeatured.configValue as HomeFeaturedConfig);
    }

    if (contactChannels) {
      contactChannelsForm.value = cloneValue(contactChannels.configValue as ContactChannelConfig[]);
    }

    if (footerConfig) {
      assignObject(footerForm, footerConfig.configValue as FooterConfig);
    }

    if (globalCta) {
      assignObject(globalCtaForm, globalCta.configValue as GlobalCtaConfig);
    }

    if (featuredSlots) {
      assignObject(featuredSlotsForm, featuredSlots.configValue as FeaturedContentSlotsConfig);
    }
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "加载站点配置失败");
  } finally {
    loading.value = false;
  }
}

async function saveConfig<T>(key: SiteConfigKey, value: T, successText: string) {
  savingKey.value = key;
  try {
    const record = await updateSiteConfig(key, value);
    lastUpdatedMap[key] = record.updatedAt;
    ElMessage.success(successText);
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "保存配置失败");
  } finally {
    savingKey.value = "";
  }
}

function addContactChannel() {
  contactChannelsForm.value.push({
    title: "",
    desc: "",
    account: "",
    buttonText: "",
    link: "#"
  });
}

function removeContactChannel(index: number) {
  contactChannelsForm.value.splice(index, 1);
}

function addFooterNavItem() {
  footerForm.navLinks.push({
    label: "",
    path: "/"
  });
}

function removeFooterNavItem(index: number) {
  footerForm.navLinks.splice(index, 1);
}

onMounted(() => {
  void loadConfigs();
});
</script>

<template>
  <div class="view">
    <div class="toolbar">
      <div>
        <h2>站点配置</h2>
        <p>集中维护首页 Banner、首页指标文案和 Contact 联系方式配置。</p>
      </div>
      <el-button @click="loadConfigs" :loading="loading">刷新配置</el-button>
    </div>

    <el-card shadow="never" v-loading="loading">
      <template #header>
        <div class="section-header">
          <div>
            <strong>首页 Banner</strong>
            <p>控制首页首屏标题、副标题和两个主 CTA。</p>
          </div>
          <div class="section-meta">最近更新：{{ lastUpdatedMap.home_banner || "-" }}</div>
        </div>
      </template>

      <el-form label-position="top" class="config-form">
        <div class="grid-two">
          <el-form-item label="主标题">
            <el-input v-model="homeBannerForm.titleText" />
          </el-form-item>
          <el-form-item label="状态标签">
            <el-input v-model="homeBannerForm.statusLabel" />
          </el-form-item>
        </div>

        <el-form-item label="副标题">
          <el-input v-model="homeBannerForm.subtitle" />
        </el-form-item>

        <div class="grid-two">
          <el-form-item label="主按钮文案">
            <el-input v-model="homeBannerForm.primaryCtaLabel" />
          </el-form-item>
          <el-form-item label="主按钮路径">
            <el-input v-model="homeBannerForm.primaryCtaPath" placeholder="/builders" />
          </el-form-item>
        </div>

        <div class="grid-two">
          <el-form-item label="次按钮文案">
            <el-input v-model="homeBannerForm.secondaryCtaLabel" />
          </el-form-item>
          <el-form-item label="次按钮路径">
            <el-input v-model="homeBannerForm.secondaryCtaPath" placeholder="/events" />
          </el-form-item>
        </div>

        <div class="actions">
          <el-button
            type="primary"
            :loading="savingKey === 'home_banner'"
            @click="saveConfig('home_banner', homeBannerForm, '首页 Banner 已更新')"
          >
            保存 Banner
          </el-button>
        </div>
      </el-form>
    </el-card>

    <el-card shadow="never">
      <template #header>
        <div class="section-header">
          <div>
            <strong>首页精选位</strong>
            <p>维护首页指标数字和几个核心区块的说明文案。</p>
          </div>
          <div class="section-meta">最近更新：{{ lastUpdatedMap.home_featured || "-" }}</div>
        </div>
      </template>

      <el-form label-position="top" class="config-form">
        <div class="grid-four">
          <el-form-item label="社区成员数">
            <el-input v-model="homeFeaturedForm.communityCount" />
          </el-form-item>
          <el-form-item label="社区成员后缀">
            <el-input v-model="homeFeaturedForm.communityCountSuffix" />
          </el-form-item>
          <el-form-item label="活动场次">
            <el-input v-model="homeFeaturedForm.eventCount" />
          </el-form-item>
          <el-form-item label="活动后缀">
            <el-input v-model="homeFeaturedForm.eventCountSuffix" />
          </el-form-item>
        </div>

        <el-form-item label="活动区描述">
          <el-input v-model="homeFeaturedForm.eventsDescription" />
        </el-form-item>
        <el-form-item label="Builder 区描述">
          <el-input v-model="homeFeaturedForm.buildersDescription" />
        </el-form-item>
        <el-form-item label="Insights 区描述">
          <el-input v-model="homeFeaturedForm.insightsDescription" />
        </el-form-item>

        <div class="actions">
          <el-button
            type="primary"
            :loading="savingKey === 'home_featured'"
            @click="saveConfig('home_featured', homeFeaturedForm, '首页精选配置已更新')"
          >
            保存首页精选位
          </el-button>
        </div>
      </el-form>
    </el-card>

    <el-card shadow="never">
      <template #header>
        <div class="section-header">
          <div>
            <strong>Contact 联系方式</strong>
            <p>维护联系页面顶部的卡片矩阵内容。</p>
          </div>
          <div class="section-meta">最近更新：{{ lastUpdatedMap.contact_channels || "-" }}</div>
        </div>
      </template>

      <div class="channel-list">
        <el-card v-for="(channel, index) in contactChannelsForm" :key="index" shadow="never" class="channel-card">
          <template #header>
            <div class="channel-header">
              <strong>联系方式 {{ index + 1 }}</strong>
              <el-button link type="danger" @click="removeContactChannel(index)">删除</el-button>
            </div>
          </template>

          <el-form label-position="top" class="config-form">
            <div class="grid-two">
              <el-form-item label="标题">
                <el-input v-model="channel.title" />
              </el-form-item>
              <el-form-item label="按钮文案">
                <el-input v-model="channel.buttonText" />
              </el-form-item>
            </div>
            <el-form-item label="说明">
              <el-input v-model="channel.desc" type="textarea" :rows="3" />
            </el-form-item>
            <div class="grid-two">
              <el-form-item label="账号/说明">
                <el-input v-model="channel.account" />
              </el-form-item>
              <el-form-item label="链接">
                <el-input v-model="channel.link" />
              </el-form-item>
            </div>
          </el-form>
        </el-card>
      </div>

      <div class="actions between">
        <el-button @click="addContactChannel">新增联系方式</el-button>
        <el-button
          type="primary"
          :loading="savingKey === 'contact_channels'"
          @click="saveConfig('contact_channels', contactChannelsForm, 'Contact 配置已更新')"
        >
          保存 Contact 配置
        </el-button>
      </div>
    </el-card>

    <el-card shadow="never">
      <template #header>
        <div class="section-header">
          <div>
            <strong>全局 CTA</strong>
            <p>控制首页全局行动召唤区域的标题、描述和两个按钮。</p>
          </div>
          <div class="section-meta">最近更新：{{ lastUpdatedMap.global_cta || "-" }}</div>
        </div>
      </template>

      <el-form label-position="top" class="config-form">
        <div class="grid-two">
          <el-form-item label="Eyebrow">
            <el-input v-model="globalCtaForm.eyebrow" />
          </el-form-item>
          <el-form-item label="主标题">
            <el-input v-model="globalCtaForm.title" />
          </el-form-item>
        </div>

        <el-form-item label="描述">
          <el-input v-model="globalCtaForm.description" type="textarea" :rows="3" />
        </el-form-item>

        <div class="grid-two">
          <el-form-item label="主按钮文案">
            <el-input v-model="globalCtaForm.primaryLabel" />
          </el-form-item>
          <el-form-item label="主按钮路径">
            <el-input v-model="globalCtaForm.primaryPath" />
          </el-form-item>
        </div>

        <div class="grid-two">
          <el-form-item label="次按钮文案">
            <el-input v-model="globalCtaForm.secondaryLabel" />
          </el-form-item>
          <el-form-item label="次按钮路径">
            <el-input v-model="globalCtaForm.secondaryPath" />
          </el-form-item>
        </div>

        <div class="actions">
          <el-button
            type="primary"
            :loading="savingKey === 'global_cta'"
            @click="saveConfig('global_cta', globalCtaForm, '全局 CTA 已更新')"
          >
            保存全局 CTA
          </el-button>
        </div>
      </el-form>
    </el-card>

    <el-card shadow="never">
      <template #header>
        <div class="section-header">
          <div>
            <strong>首页精选位策略</strong>
            <p>控制首页三个内容区块的标题、查看全部文案和展示数量。</p>
          </div>
          <div class="section-meta">最近更新：{{ lastUpdatedMap.featured_content_slots || "-" }}</div>
        </div>
      </template>

      <el-form label-position="top" class="config-form">
        <div class="grid-three">
          <el-form-item label="活动标题">
            <el-input v-model="featuredSlotsForm.eventsTitle" />
          </el-form-item>
          <el-form-item label="活动按钮文案">
            <el-input v-model="featuredSlotsForm.eventsViewAllLabel" />
          </el-form-item>
          <el-form-item label="活动展示数量">
            <el-input-number v-model="featuredSlotsForm.eventsLimit" :min="1" :max="12" class="full-width" />
          </el-form-item>
        </div>

        <div class="grid-three">
          <el-form-item label="Builder 标题">
            <el-input v-model="featuredSlotsForm.buildersTitle" />
          </el-form-item>
          <el-form-item label="Builder 按钮文案">
            <el-input v-model="featuredSlotsForm.buildersViewAllLabel" />
          </el-form-item>
          <el-form-item label="Builder 展示数量">
            <el-input-number v-model="featuredSlotsForm.buildersLimit" :min="1" :max="12" class="full-width" />
          </el-form-item>
        </div>

        <div class="grid-three">
          <el-form-item label="资讯标题">
            <el-input v-model="featuredSlotsForm.insightsTitle" />
          </el-form-item>
          <el-form-item label="资讯按钮文案">
            <el-input v-model="featuredSlotsForm.insightsViewAllLabel" />
          </el-form-item>
          <el-form-item label="资讯展示数量">
            <el-input-number v-model="featuredSlotsForm.insightsLimit" :min="1" :max="12" class="full-width" />
          </el-form-item>
        </div>

        <div class="actions">
          <el-button
            type="primary"
            :loading="savingKey === 'featured_content_slots'"
            @click="saveConfig('featured_content_slots', featuredSlotsForm, '首页精选位策略已更新')"
          >
            保存首页精选位策略
          </el-button>
        </div>
      </el-form>
    </el-card>

    <el-card shadow="never">
      <template #header>
        <div class="section-header">
          <div>
            <strong>页脚配置</strong>
            <p>统一维护首页页脚 slogan、导航链接和版权信息。</p>
          </div>
          <div class="section-meta">最近更新：{{ lastUpdatedMap.footer_config || "-" }}</div>
        </div>
      </template>

      <el-form label-position="top" class="config-form">
        <el-form-item label="页脚标语">
          <el-input v-model="footerForm.slogan" />
        </el-form-item>
        <el-form-item label="版权文案">
          <el-input v-model="footerForm.legalText" />
        </el-form-item>

        <div class="channel-list">
          <el-card
            v-for="(item, index) in footerForm.navLinks"
            :key="`${item.label}-${index}`"
            shadow="never"
            class="channel-card"
          >
            <template #header>
              <div class="channel-header">
                <strong>页脚链接 {{ index + 1 }}</strong>
                <el-button link type="danger" @click="removeFooterNavItem(index)">删除</el-button>
              </div>
            </template>

            <div class="grid-two">
              <el-form-item label="名称">
                <el-input v-model="item.label" />
              </el-form-item>
              <el-form-item label="路径">
                <el-input v-model="item.path" />
              </el-form-item>
            </div>
          </el-card>
        </div>

        <div class="actions between">
          <el-button @click="addFooterNavItem">新增页脚链接</el-button>
          <el-button
            type="primary"
            :loading="savingKey === 'footer_config'"
            @click="saveConfig('footer_config', footerForm, '页脚配置已更新')"
          >
            保存页脚配置
          </el-button>
        </div>
      </el-form>
    </el-card>
  </div>
</template>

<style scoped>
.view {
  display: grid;
  gap: 20px;
}

.toolbar,
.section-header,
.channel-header,
.actions.between {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
}

.toolbar {
  align-items: flex-start;
}

h2 {
  margin: 0;
  font-size: 28px;
}

p {
  margin: 8px 0 0;
  color: rgba(17, 24, 39, 0.6);
}

.section-meta {
  color: rgba(17, 24, 39, 0.52);
  font-size: 13px;
}

.config-form {
  display: grid;
  gap: 8px;
}

.grid-two,
.grid-three,
.grid-four {
  display: grid;
  gap: 16px;
}

.grid-two {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.grid-three {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.grid-four {
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 8px;
}

.channel-list {
  display: grid;
  gap: 16px;
}

.full-width {
  width: 100%;
}

.channel-card {
  border: 1px solid #eef2f7;
}

@media (max-width: 960px) {
  .grid-two,
  .grid-four {
    grid-template-columns: 1fr;
  }

  .toolbar,
  .section-header,
  .channel-header,
  .actions.between {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
