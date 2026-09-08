<script setup lang="ts">
// A picker for a small master-data list (departemen, lokasi) that can also add
// a new entry or rename the selected one without leaving the form. Sending HR
// to the Master Data screen mid-way through creating an employee meant losing
// whatever they had already typed.
import { CheckOutlined, CloseOutlined, EditOutlined, PlusOutlined } from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'
import { computed, ref } from 'vue'

const props = defineProps<{
  options: { label: string; value: string }[]
  placeholder?: string
  entityLabel: string
  create: (name: string) => Promise<{ id: string; name: string }>
  update: (id: string, name: string) => Promise<void>
}>()

const model = defineModel<string | undefined>()
const emit = defineEmits<{ changed: [] }>()

type Mode = 'idle' | 'add' | 'edit'
const mode = ref<Mode>('idle')
const draft = ref('')
const busy = ref(false)

const selectedLabel = computed(
  () => props.options.find((o) => o.value === model.value)?.label ?? '',
)

function startAdd() {
  draft.value = ''
  mode.value = 'add'
}

function startEdit() {
  if (!model.value) return
  draft.value = selectedLabel.value
  mode.value = 'edit'
}

function cancel() {
  mode.value = 'idle'
  draft.value = ''
}

async function save() {
  const name = draft.value.trim()
  if (!name) {
    message.error(`Nama ${props.entityLabel} wajib diisi`)
    return
  }
  busy.value = true
  try {
    if (mode.value === 'add') {
      const created = await props.create(name)
      model.value = created.id
      message.success(`${props.entityLabel} "${name}" ditambahkan`)
    } else if (model.value) {
      await props.update(model.value, name)
      message.success(`${props.entityLabel} diubah jadi "${name}"`)
    }
    emit('changed')
    cancel()
  } catch (err: any) {
    message.error(err?.response?.data?.error ?? `Gagal menyimpan ${props.entityLabel}`)
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <div class="inline-master">
    <div v-if="mode === 'idle'" class="row">
      <a-select
        v-model:value="model"
        :options="options"
        :placeholder="placeholder"
        allow-clear
        show-search
        option-filter-prop="label"
        class="grow"
      />
      <a-tooltip :title="`Tambah ${entityLabel} baru`">
        <a-button @click="startAdd"><template #icon><PlusOutlined /></template></a-button>
      </a-tooltip>
      <a-tooltip :title="model ? `Ubah nama ${entityLabel} terpilih` : `Pilih ${entityLabel} dulu untuk mengubahnya`">
        <a-button :disabled="!model" @click="startEdit">
          <template #icon><EditOutlined /></template>
        </a-button>
      </a-tooltip>
    </div>

    <div v-else class="row">
      <a-input
        v-model:value="draft"
        class="grow"
        :placeholder="mode === 'add' ? `Nama ${entityLabel} baru` : `Nama ${entityLabel}`"
        autofocus
        @press-enter="save"
      />
      <a-button type="primary" :loading="busy" @click="save">
        <template #icon><CheckOutlined /></template>
      </a-button>
      <a-button :disabled="busy" @click="cancel">
        <template #icon><CloseOutlined /></template>
      </a-button>
    </div>
    <p v-if="mode === 'add'" class="hint">{{ entityLabel }} baru akan langsung terpilih setelah disimpan.</p>
  </div>
</template>

<style scoped>
.row {
  display: flex;
  gap: 8px;
}
.grow {
  flex: 1;
  min-width: 0;
}
.hint {
  margin: 6px 0 0;
  color: #94a3b8;
  font-size: 12px;
}
</style>
