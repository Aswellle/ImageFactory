<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useProjectStore } from '@/stores/project'

const router = useRouter()
const projectStore = useProjectStore()

const showCreate = ref(false)
const newName = ref('')
const newDesc = ref('')
const searchQuery = ref('')

const filteredProjects = computed(() => {
  if (!searchQuery.value) return projectStore.projects
  const q = searchQuery.value.toLowerCase()
  return projectStore.projects.filter(
    (p) => p.name.toLowerCase().includes(q) || p.description?.toLowerCase().includes(q),
  )
})

onMounted(() => {
  projectStore.fetchProjects()
})

async function createProject() {
  if (!newName.value.trim()) return
  try {
    await projectStore.createProject(newName.value.trim(), newDesc.value.trim() || undefined)
    newName.value = ''
    newDesc.value = ''
    showCreate.value = false
  } catch {
    // error in store
  }
}

function openProject(id: number) {
  router.push(`/projects/${id}`)
}

function formatDate(dateStr: string): string {
  const d = new Date(dateStr)
  return d.toLocaleDateString('zh-CN', { month: 'short', day: 'numeric', year: 'numeric' })
}

const creating = ref(false)

async function handleCreate() {
  if (!newName.value.trim()) return
  creating.value = true
  try {
    await createProject()
  } finally {
    creating.value = false
  }
}
</script>

<template>
  <div class="studio-theme">
    <div class="projects-studio">
      <!-- Sidebar -->
      <aside class="studio-sidebar">
        <div class="studio-sidebar-logo">
          <svg width="18" height="18" viewBox="0 0 32 32" fill="none">
            <path d="M8 22V10l8 6-8 6zM16 10l8 6-8 6V10z" fill="#0C0A09" />
          </svg>
        </div>
        <nav class="studio-sidebar-nav">
          <RouterLink to="/app" class="studio-nav-item" title="Generate">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              <rect x="3" y="3" width="18" height="18" rx="2" />
              <circle cx="8.5" cy="8.5" r="1.5" />
              <path d="M21 15l-5-5L5 21" />
            </svg>
          </RouterLink>
          <RouterLink to="/app/gallery" class="studio-nav-item" title="Gallery">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              <rect x="3" y="3" width="18" height="18" rx="2" />
              <path d="M3 15l4-4a2 2 0 012.8 0L14 15" />
              <path d="M14 13l1-1a2 2 0 012.8 0L21 15" />
            </svg>
          </RouterLink>
          <RouterLink to="/app/projects" class="studio-nav-item active" title="Projects">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              <path d="M22 19a2 2 0 01-2 2H4a2 2 0 01-2-2V5a2 2 0 012-2h5l2 3h9a2 2 0 012 2z" />
            </svg>
          </RouterLink>
          <RouterLink to="/app/api" class="studio-nav-item" title="API">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              <path d="M16 18l6-6-6-6M8 6l-6 6 6 6" />
            </svg>
          </RouterLink>
        </nav>
      </aside>

      <!-- Main -->
      <div class="projects-main">
        <!-- Top bar -->
        <header class="projects-topbar">
          <div class="projects-topbar-left">
            <span class="projects-topbar-title">Projects</span>
            <span class="projects-topbar-count">{{ filteredProjects.length }} projects</span>
          </div>
          <div class="projects-topbar-right">
            <div class="projects-search">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><path d="M21 21l-4.35-4.35"/></svg>
              <input v-model="searchQuery" type="text" placeholder="Search projects..." />
            </div>
            <button class="projects-new-btn" @click="showCreate = true">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M12 5v14M5 12h14"/></svg>
              New Project
            </button>
          </div>
        </header>

        <!-- Content -->
        <div class="projects-content">
          <!-- Loading -->
          <div v-if="projectStore.loading" class="projects-skeleton-grid">
            <div v-for="n in 6" :key="n" class="projects-skeleton-card">
              <div class="projects-skeleton-previews">
                <div v-for="i in 9" :key="i" class="projects-skeleton-preview"></div>
              </div>
              <div class="projects-skeleton-info">
                <div class="projects-skeleton-line"></div>
                <div class="projects-skeleton-line short"></div>
              </div>
            </div>
          </div>

          <!-- Empty -->
          <div v-else-if="filteredProjects.length === 0" class="projects-empty">
            <div class="projects-empty-icon">
              <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="text-[var(--studio-text-muted)]">
                <path d="M22 19a2 2 0 01-2 2H4a2 2 0 01-2-2V5a2 2 0 012-2h5l2 3h9a2 2 0 012 2z" />
              </svg>
            </div>
            <p class="projects-empty-title">{{ searchQuery ? 'No matching projects' : 'No projects yet' }}</p>
            <p class="projects-empty-desc">{{ searchQuery ? 'Try adjusting your search' : 'Create a project to organize your images into collections' }}</p>
            <button v-if="!searchQuery" class="projects-empty-btn" @click="showCreate = true">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M12 5v14M5 12h14"/></svg>
              Create your first project
            </button>
            <button v-else class="projects-empty-btn" @click="searchQuery = ''">
              Clear search
            </button>
          </div>

          <!-- Contact Sheet Grid -->
          <div v-else class="contact-sheet-grid">
            <div
              v-for="project in filteredProjects"
              :key="project.id"
              class="contact-sheet"
              @click="openProject(project.id)"
            >
              <!-- Mini preview grid (3x3) - shows empty film frames -->
              <div class="contact-sheet-previews">
                <template v-for="_ in 9" :key="_">
                  <div class="contact-sheet-preview">
                    <div class="contact-sheet-preview-empty"></div>
                  </div>
                </template>
              </div>
              <!-- Info -->
              <div class="contact-sheet-info">
                <h3 class="contact-sheet-name">{{ project.name }}</h3>
                <p v-if="project.description" class="contact-sheet-desc">{{ project.description }}</p>
                <div class="contact-sheet-meta">
                  <span>Project</span>
                  <span class="contact-sheet-meta-dot"></span>
                  <span>{{ formatDate(project.created_at) }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Create Modal -->
      <div v-if="showCreate" class="projects-modal-overlay" @click.self="showCreate = false">
        <div class="projects-modal">
          <h2 class="projects-modal-title">New Project</h2>
          <div class="projects-modal-field">
            <label class="projects-modal-label">Name</label>
            <input
              v-model="newName"
              class="projects-modal-input"
              placeholder="e.g. Summer Campaign"
              autofocus
              @keydown.enter="handleCreate"
            />
          </div>
          <div class="projects-modal-field">
            <label class="projects-modal-label">Description (optional)</label>
            <textarea
              v-model="newDesc"
              class="projects-modal-input projects-modal-textarea"
              placeholder="What's this project about..."
            ></textarea>
          </div>
          <div class="projects-modal-actions">
            <button class="projects-modal-btn" @click="showCreate = false">Cancel</button>
            <button class="projects-modal-btn primary" :disabled="!newName.trim() || creating" @click="handleCreate">
              {{ creating ? 'Creating...' : 'Create' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
