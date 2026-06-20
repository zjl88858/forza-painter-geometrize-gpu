package gpu

/*
#cgo windows CFLAGS: -IC:/VulkanSDK/1.4.350.0/Include
#cgo windows LDFLAGS: -LC:/VulkanSDK/1.4.350.0/Lib -lvulkan-1
#include <stdlib.h>
#include <string.h>
#include <stdint.h>
#include <vulkan/vulkan.h>

static VkResult rawVkCreateInstance(const VkInstanceCreateInfo* pCreateInfo, const VkAllocationCallbacks* pAllocator, VkInstance* pInstance) {
	return vkCreateInstance(pCreateInfo, pAllocator, pInstance);
}

static VkResult rawVkCreateInstanceSimple(const char* appName, const char* engineName, VkInstance* pInstance) {
	VkApplicationInfo appInfo = {
		.sType = VK_STRUCTURE_TYPE_APPLICATION_INFO,
		.pApplicationName = appName,
		.applicationVersion = VK_MAKE_VERSION(1, 1, 0),
		.pEngineName = engineName,
		.engineVersion = VK_MAKE_VERSION(1, 0, 0),
		.apiVersion = VK_MAKE_VERSION(1, 2, 0),
	};
	VkInstanceCreateInfo ci = {
		.sType = VK_STRUCTURE_TYPE_INSTANCE_CREATE_INFO,
		.pApplicationInfo = &appInfo,
	};
	return vkCreateInstance(&ci, NULL, pInstance);
}

static void rawVkDestroyInstance(VkInstance instance, const VkAllocationCallbacks* pAllocator) {
	vkDestroyInstance(instance, pAllocator);
}

static VkResult rawVkEnumeratePhysicalDevices(VkInstance instance, uint32_t* pPhysicalDeviceCount, VkPhysicalDevice* pPhysicalDevices) {
	return vkEnumeratePhysicalDevices(instance, pPhysicalDeviceCount, pPhysicalDevices);
}

static void rawVkGetPhysicalDeviceProperties(VkPhysicalDevice physicalDevice, VkPhysicalDeviceProperties* pProperties) {
	vkGetPhysicalDeviceProperties(physicalDevice, pProperties);
}

static void rawVkGetPhysicalDeviceQueueFamilyProperties(VkPhysicalDevice physicalDevice, uint32_t* pQueueFamilyPropertyCount, VkQueueFamilyProperties* pQueueFamilyProperties) {
	vkGetPhysicalDeviceQueueFamilyProperties(physicalDevice, pQueueFamilyPropertyCount, pQueueFamilyProperties);
}

static void rawVkGetPhysicalDeviceMemoryProperties(VkPhysicalDevice physicalDevice, VkPhysicalDeviceMemoryProperties* pMemoryProperties) {
	vkGetPhysicalDeviceMemoryProperties(physicalDevice, pMemoryProperties);
}

static VkResult rawVkCreateDevice(VkPhysicalDevice physicalDevice, const VkDeviceCreateInfo* pCreateInfo, const VkAllocationCallbacks* pAllocator, VkDevice* pDevice) {
	return vkCreateDevice(physicalDevice, pCreateInfo, pAllocator, pDevice);
}

static VkResult rawVkCreateDeviceSimple(VkPhysicalDevice physicalDevice, uint32_t queueFamilyIndex, VkDevice* pDevice, VkQueue* pQueue) {
	float priority = 1.0f;
	VkDeviceQueueCreateInfo qci = {
		.sType = VK_STRUCTURE_TYPE_DEVICE_QUEUE_CREATE_INFO,
		.queueFamilyIndex = queueFamilyIndex,
		.queueCount = 1,
		.pQueuePriorities = &priority,
	};
	VkPhysicalDeviceFeatures features = {0};
	VkDeviceCreateInfo devCI = {
		.sType = VK_STRUCTURE_TYPE_DEVICE_CREATE_INFO,
		.queueCreateInfoCount = 1,
		.pQueueCreateInfos = &qci,
		.pEnabledFeatures = &features,
	};
	VkResult res = vkCreateDevice(physicalDevice, &devCI, NULL, pDevice);
	if (res == VK_SUCCESS) {
		vkGetDeviceQueue(*pDevice, queueFamilyIndex, 0, pQueue);
	}
	return res;
}

static void rawVkDestroyDevice(VkDevice device, const VkAllocationCallbacks* pAllocator) {
	vkDestroyDevice(device, pAllocator);
}

static void rawVkGetDeviceQueue(VkDevice device, uint32_t queueFamilyIndex, uint32_t queueIndex, VkQueue* pQueue) {
	vkGetDeviceQueue(device, queueFamilyIndex, queueIndex, pQueue);
}

static VkResult rawVkCreateCommandPool(VkDevice device, const VkCommandPoolCreateInfo* pCreateInfo, const VkAllocationCallbacks* pAllocator, VkCommandPool* pCommandPool) {
	return vkCreateCommandPool(device, pCreateInfo, pAllocator, pCommandPool);
}

static VkResult rawVkCreateCommandPoolSimple(VkDevice device, uint32_t queueFamilyIndex, VkCommandPool* pCommandPool) {
	VkCommandPoolCreateInfo ci = {
		.sType = VK_STRUCTURE_TYPE_COMMAND_POOL_CREATE_INFO,
		.queueFamilyIndex = queueFamilyIndex,
		.flags = VK_COMMAND_POOL_CREATE_RESET_COMMAND_BUFFER_BIT,
	};
	return vkCreateCommandPool(device, &ci, NULL, pCommandPool);
}

static void rawVkDestroyCommandPool(VkDevice device, VkCommandPool commandPool, const VkAllocationCallbacks* pAllocator) {
	vkDestroyCommandPool(device, commandPool, pAllocator);
}

static VkResult rawVkAllocateMemory(VkDevice device, const VkMemoryAllocateInfo* pAllocateInfo, const VkAllocationCallbacks* pAllocator, VkDeviceMemory* pMemory) {
	return vkAllocateMemory(device, pAllocateInfo, pAllocator, pMemory);
}

static void rawVkFreeMemory(VkDevice device, VkDeviceMemory memory, const VkAllocationCallbacks* pAllocator) {
	vkFreeMemory(device, memory, pAllocator);
}

static VkResult rawVkCreateBuffer(VkDevice device, const VkBufferCreateInfo* pCreateInfo, const VkAllocationCallbacks* pAllocator, VkBuffer* pBuffer) {
	return vkCreateBuffer(device, pCreateInfo, pAllocator, pBuffer);
}

static void rawVkDestroyBuffer(VkDevice device, VkBuffer buffer, const VkAllocationCallbacks* pAllocator) {
	vkDestroyBuffer(device, buffer, pAllocator);
}

static void rawVkGetBufferMemoryRequirements(VkDevice device, VkBuffer buffer, VkMemoryRequirements* pMemoryRequirements) {
	vkGetBufferMemoryRequirements(device, buffer, pMemoryRequirements);
}

static VkResult rawVkBindBufferMemory(VkDevice device, VkBuffer buffer, VkDeviceMemory memory, VkDeviceSize memoryOffset) {
	return vkBindBufferMemory(device, buffer, memory, memoryOffset);
}

static VkResult rawVkMapMemory(VkDevice device, VkDeviceMemory memory, VkDeviceSize offset, VkDeviceSize size, VkMemoryMapFlags flags, void** ppData) {
	return vkMapMemory(device, memory, offset, size, flags, ppData);
}

static void rawVkUnmapMemory(VkDevice device, VkDeviceMemory memory) {
	vkUnmapMemory(device, memory);
}

static VkResult rawVkCreateDescriptorSetLayout(VkDevice device, const VkDescriptorSetLayoutCreateInfo* pCreateInfo, const VkAllocationCallbacks* pAllocator, VkDescriptorSetLayout* pSetLayout) {
	return vkCreateDescriptorSetLayout(device, pCreateInfo, pAllocator, pSetLayout);
}

static VkResult rawVkCreateDescriptorSetLayoutSimple(VkDevice device, VkDescriptorSetLayout* pSetLayout) {
	VkDescriptorSetLayoutBinding bindings[5] = {
		{0, VK_DESCRIPTOR_TYPE_STORAGE_BUFFER, 1, VK_SHADER_STAGE_COMPUTE_BIT, NULL},
		{1, VK_DESCRIPTOR_TYPE_STORAGE_BUFFER, 1, VK_SHADER_STAGE_COMPUTE_BIT, NULL},
		{2, VK_DESCRIPTOR_TYPE_STORAGE_BUFFER, 1, VK_SHADER_STAGE_COMPUTE_BIT, NULL},
		{3, VK_DESCRIPTOR_TYPE_STORAGE_BUFFER, 1, VK_SHADER_STAGE_COMPUTE_BIT, NULL},
		{4, VK_DESCRIPTOR_TYPE_STORAGE_BUFFER, 1, VK_SHADER_STAGE_COMPUTE_BIT, NULL},
	};
	VkDescriptorSetLayoutCreateInfo ci = {
		.sType = VK_STRUCTURE_TYPE_DESCRIPTOR_SET_LAYOUT_CREATE_INFO,
		.bindingCount = 5,
		.pBindings = bindings,
	};
	return vkCreateDescriptorSetLayout(device, &ci, NULL, pSetLayout);
}

static void rawVkDestroyDescriptorSetLayout(VkDevice device, VkDescriptorSetLayout descriptorSetLayout, const VkAllocationCallbacks* pAllocator) {
	vkDestroyDescriptorSetLayout(device, descriptorSetLayout, pAllocator);
}

static VkResult rawVkCreateDescriptorPool(VkDevice device, const VkDescriptorPoolCreateInfo* pCreateInfo, const VkAllocationCallbacks* pAllocator, VkDescriptorPool* pDescriptorPool) {
	return vkCreateDescriptorPool(device, pCreateInfo, pAllocator, pDescriptorPool);
}

static VkResult rawVkCreateDescriptorPoolSimple(VkDevice device, VkDescriptorPool* pDescriptorPool) {
	VkDescriptorPoolSize sizes[1] = {
		{VK_DESCRIPTOR_TYPE_STORAGE_BUFFER, (uint32_t)(5 * 6)},
	};
	VkDescriptorPoolCreateInfo ci = {
		.sType = VK_STRUCTURE_TYPE_DESCRIPTOR_POOL_CREATE_INFO,
		.maxSets = 6,
		.poolSizeCount = 1,
		.pPoolSizes = sizes,
	};
	return vkCreateDescriptorPool(device, &ci, NULL, pDescriptorPool);
}

static VkResult rawVkCreateDescriptorPoolSized(VkDevice device, uint32_t maxSets, uint32_t storageCount, VkDescriptorPool* pDescriptorPool) {
	VkDescriptorPoolSize sizes[1] = {
		{VK_DESCRIPTOR_TYPE_STORAGE_BUFFER, storageCount},
	};
	VkDescriptorPoolCreateInfo ci = {
		.sType = VK_STRUCTURE_TYPE_DESCRIPTOR_POOL_CREATE_INFO,
		.maxSets = maxSets,
		.poolSizeCount = 1,
		.pPoolSizes = sizes,
	};
	return vkCreateDescriptorPool(device, &ci, NULL, pDescriptorPool);
}

static void rawVkDestroyDescriptorPool(VkDevice device, VkDescriptorPool descriptorPool, const VkAllocationCallbacks* pAllocator) {
	vkDestroyDescriptorPool(device, descriptorPool, pAllocator);
}

static VkResult rawVkAllocateDescriptorSets(VkDevice device, const VkDescriptorSetAllocateInfo* pAllocateInfo, VkDescriptorSet* pDescriptorSets) {
	return vkAllocateDescriptorSets(device, pAllocateInfo, pDescriptorSets);
}

static VkResult rawVkAllocateDescriptorSetsSimple(VkDevice device, VkDescriptorPool descriptorPool, VkDescriptorSetLayout layout, uint32_t count, VkDescriptorSet* pDescriptorSets) {
	VkDescriptorSetLayout layouts[3] = {layout, layout, layout};
	VkDescriptorSetAllocateInfo ci = {
		.sType = VK_STRUCTURE_TYPE_DESCRIPTOR_SET_ALLOCATE_INFO,
		.descriptorPool = descriptorPool,
		.descriptorSetCount = count,
		.pSetLayouts = layouts,
	};
	return vkAllocateDescriptorSets(device, &ci, pDescriptorSets);
}

static void rawVkUpdateDescriptorSets(VkDevice device, uint32_t descriptorWriteCount, const VkWriteDescriptorSet* pDescriptorWrites) {
	vkUpdateDescriptorSets(device, descriptorWriteCount, pDescriptorWrites, 0, NULL);
}

static void rawVkUpdateDescriptorSetStorageBuffer(VkDevice device, VkDescriptorSet set, uint32_t binding, VkBuffer buffer) {
	VkDescriptorBufferInfo info = {buffer, 0, VK_WHOLE_SIZE};
	VkWriteDescriptorSet write = {
		.sType = VK_STRUCTURE_TYPE_WRITE_DESCRIPTOR_SET,
		.dstSet = set,
		.dstBinding = binding,
		.dstArrayElement = 0,
		.descriptorCount = 1,
		.descriptorType = VK_DESCRIPTOR_TYPE_STORAGE_BUFFER,
		.pBufferInfo = &info,
	};
	vkUpdateDescriptorSets(device, 1, &write, 0, NULL);
}

static VkResult rawVkCreateShaderModule(VkDevice device, const VkShaderModuleCreateInfo* pCreateInfo, const VkAllocationCallbacks* pAllocator, VkShaderModule* pShaderModule) {
	return vkCreateShaderModule(device, pCreateInfo, pAllocator, pShaderModule);
}

static VkResult rawVkCreateShaderModuleSimple(VkDevice device, const uint32_t* code, size_t codeSize, VkShaderModule* pShaderModule) {
	VkShaderModuleCreateInfo ci = {
		.sType = VK_STRUCTURE_TYPE_SHADER_MODULE_CREATE_INFO,
		.codeSize = codeSize,
		.pCode = code,
	};
	return vkCreateShaderModule(device, &ci, NULL, pShaderModule);
}

static void rawVkDestroyShaderModule(VkDevice device, VkShaderModule shaderModule, const VkAllocationCallbacks* pAllocator) {
	vkDestroyShaderModule(device, shaderModule, pAllocator);
}

static VkResult rawVkCreatePipelineLayout(VkDevice device, const VkPipelineLayoutCreateInfo* pCreateInfo, const VkAllocationCallbacks* pAllocator, VkPipelineLayout* pPipelineLayout) {
	return vkCreatePipelineLayout(device, pCreateInfo, pAllocator, pPipelineLayout);
}

static VkResult rawVkCreatePipelineLayoutSimple(VkDevice device, VkDescriptorSetLayout setLayout, VkPipelineLayout* pPipelineLayout) {
	VkPushConstantRange pcRange = {VK_SHADER_STAGE_COMPUTE_BIT, 0, 64};
	VkPipelineLayoutCreateInfo ci = {
		.sType = VK_STRUCTURE_TYPE_PIPELINE_LAYOUT_CREATE_INFO,
		.setLayoutCount = 1,
		.pSetLayouts = &setLayout,
		.pushConstantRangeCount = 1,
		.pPushConstantRanges = &pcRange,
	};
	return vkCreatePipelineLayout(device, &ci, NULL, pPipelineLayout);
}

static VkResult rawVkCreateComputePipelineSimple2(VkDevice device, VkShaderModule module, VkPipelineLayout layout, VkPipeline* pPipeline) {
	VkPipelineShaderStageCreateInfo stage = {
		.sType = VK_STRUCTURE_TYPE_PIPELINE_SHADER_STAGE_CREATE_INFO,
		.stage = VK_SHADER_STAGE_COMPUTE_BIT,
		.module = module,
		.pName = "main",
	};
	VkComputePipelineCreateInfo ci = {
		.sType = VK_STRUCTURE_TYPE_COMPUTE_PIPELINE_CREATE_INFO,
		.stage = stage,
		.layout = layout,
	};
	return vkCreateComputePipelines(device, NULL, 1, &ci, NULL, pPipeline);
}

static void rawVkDestroyPipelineLayout(VkDevice device, VkPipelineLayout pipelineLayout, const VkAllocationCallbacks* pAllocator) {
	vkDestroyPipelineLayout(device, pipelineLayout, pAllocator);
}

static VkResult rawVkCreateComputePipelines(VkDevice device, VkPipelineCache pipelineCache, uint32_t createInfoCount, const VkComputePipelineCreateInfo* pCreateInfos, const VkAllocationCallbacks* pAllocator, VkPipeline* pPipelines) {
	return vkCreateComputePipelines(device, pipelineCache, createInfoCount, pCreateInfos, pAllocator, pPipelines);
}

static VkResult rawVkCreateComputePipelineSimple(VkDevice device, VkShaderModule module, VkPipelineLayout layout, VkPipeline* pPipeline) {
	VkPipelineShaderStageCreateInfo stage = {
		.sType = VK_STRUCTURE_TYPE_PIPELINE_SHADER_STAGE_CREATE_INFO,
		.stage = VK_SHADER_STAGE_COMPUTE_BIT,
		.module = module,
		.pName = "main",
	};
	VkComputePipelineCreateInfo ci = {
		.sType = VK_STRUCTURE_TYPE_COMPUTE_PIPELINE_CREATE_INFO,
		.stage = stage,
		.layout = layout,
	};
	return vkCreateComputePipelines(device, NULL, 1, &ci, NULL, pPipeline);
}

static void rawVkDestroyPipeline(VkDevice device, VkPipeline pipeline, const VkAllocationCallbacks* pAllocator) {
	vkDestroyPipeline(device, pipeline, pAllocator);
}

static VkResult rawVkCreateFence(VkDevice device, const VkFenceCreateInfo* pCreateInfo, const VkAllocationCallbacks* pAllocator, VkFence* pFence) {
	return vkCreateFence(device, pCreateInfo, pAllocator, pFence);
}

static void rawVkDestroyFence(VkDevice device, VkFence fence, const VkAllocationCallbacks* pAllocator) {
	vkDestroyFence(device, fence, pAllocator);
}

static VkResult rawVkResetFences(VkDevice device, uint32_t fenceCount, const VkFence* pFences) {
	return vkResetFences(device, fenceCount, pFences);
}

static VkResult rawVkWaitForFences(VkDevice device, uint32_t fenceCount, const VkFence* pFences, VkBool32 waitAll, uint64_t timeout) {
	return vkWaitForFences(device, fenceCount, pFences, waitAll, timeout);
}

static VkResult rawVkCreateSemaphore(VkDevice device, const VkSemaphoreCreateInfo* pCreateInfo, const VkAllocationCallbacks* pAllocator, VkSemaphore* pSemaphore) {
	return vkCreateSemaphore(device, pCreateInfo, pAllocator, pSemaphore);
}

static void rawVkDestroySemaphore(VkDevice device, VkSemaphore semaphore, const VkAllocationCallbacks* pAllocator) {
	vkDestroySemaphore(device, semaphore, pAllocator);
}

static VkResult rawVkCreateCommandBuffer(VkDevice device, const VkCommandBufferAllocateInfo* pAllocateInfo, VkCommandBuffer* pCommandBuffers) {
	return vkAllocateCommandBuffers(device, pAllocateInfo, pCommandBuffers);
}

static void rawVkFreeCommandBuffers(VkDevice device, VkCommandPool commandPool, uint32_t commandBufferCount, const VkCommandBuffer* pCommandBuffers) {
	vkFreeCommandBuffers(device, commandPool, commandBufferCount, pCommandBuffers);
}

static VkResult rawVkBeginCommandBuffer(VkCommandBuffer commandBuffer, const VkCommandBufferBeginInfo* pBeginInfo) {
	return vkBeginCommandBuffer(commandBuffer, pBeginInfo);
}

static VkResult rawVkEndCommandBuffer(VkCommandBuffer commandBuffer) {
	return vkEndCommandBuffer(commandBuffer);
}

static void rawVkCmdBindPipeline(VkCommandBuffer commandBuffer, VkPipelineBindPoint pipelineBindPoint, VkPipeline pipeline) {
	vkCmdBindPipeline(commandBuffer, pipelineBindPoint, pipeline);
}

static void rawVkCmdBindDescriptorSets(VkCommandBuffer commandBuffer, VkPipelineBindPoint pipelineBindPoint, VkPipelineLayout layout, uint32_t firstSet, uint32_t descriptorSetCount, const VkDescriptorSet* pDescriptorSets) {
	vkCmdBindDescriptorSets(commandBuffer, pipelineBindPoint, layout, firstSet, descriptorSetCount, pDescriptorSets, 0, NULL);
}

static void rawVkCmdPushConstants(VkCommandBuffer commandBuffer, VkPipelineLayout layout, VkShaderStageFlags stageFlags, uint32_t offset, uint32_t size, const void* pValues) {
	vkCmdPushConstants(commandBuffer, layout, stageFlags, offset, size, pValues);
}

static void rawVkCmdDispatch(VkCommandBuffer commandBuffer, uint32_t groupCountX, uint32_t groupCountY, uint32_t groupCountZ) {
	vkCmdDispatch(commandBuffer, groupCountX, groupCountY, groupCountZ);
}

static void rawVkCmdCopyBuffer(VkCommandBuffer commandBuffer, VkBuffer srcBuffer, VkBuffer dstBuffer, uint32_t regionCount, const VkBufferCopy* pRegions) {
	vkCmdCopyBuffer(commandBuffer, srcBuffer, dstBuffer, regionCount, pRegions);
}

static void rawVkCmdPipelineBarrier(VkCommandBuffer commandBuffer, VkPipelineStageFlags srcStageMask, VkPipelineStageFlags dstStageMask, VkDependencyFlags dependencyFlags, uint32_t memoryBarrierCount, const VkMemoryBarrier* pMemoryBarriers, uint32_t bufferMemoryBarrierCount, const VkBufferMemoryBarrier* pBufferMemoryBarriers, uint32_t imageMemoryBarrierCount, const VkImageMemoryBarrier* pImageMemoryBarriers) {
	vkCmdPipelineBarrier(commandBuffer, srcStageMask, dstStageMask, dependencyFlags, memoryBarrierCount, pMemoryBarriers, bufferMemoryBarrierCount, pBufferMemoryBarriers, imageMemoryBarrierCount, pImageMemoryBarriers);
}

static VkResult rawVkQueueSubmit(VkQueue queue, uint32_t submitCount, const VkSubmitInfo* pSubmits, VkFence fence) {
	return vkQueueSubmit(queue, submitCount, pSubmits, fence);
}

static VkResult rawVkQueueSubmitOne(VkQueue queue, VkCommandBuffer cmdBuf, VkFence fence) {
	VkSubmitInfo submitInfo = {
		.sType = VK_STRUCTURE_TYPE_SUBMIT_INFO,
		.commandBufferCount = 1,
		.pCommandBuffers = &cmdBuf,
	};
	return vkQueueSubmit(queue, 1, &submitInfo, fence);
}

static VkResult rawVkResetCommandBufferOne(VkCommandBuffer commandBuffer) {
	return vkResetCommandBuffer(commandBuffer, 0);
}

static VkResult rawVkWaitForFencesOne(VkDevice device, VkFence fence, uint64_t timeout) {
	return vkWaitForFences(device, 1, &fence, VK_TRUE, timeout);
}

static VkResult rawVkResetFencesOne(VkDevice device, VkFence fence) {
	return vkResetFences(device, 1, &fence);
}

static VkResult rawVkQueueWaitIdle(VkQueue queue) {
	return vkQueueWaitIdle(queue);
}

static VkResult rawVkDeviceWaitIdle(VkDevice device) {
	return vkDeviceWaitIdle(device);
}

static VkDeviceSize rawVkWholeSize() {
	return VK_WHOLE_SIZE;
}

static VkQueueFamilyProperties* allocQueueFamilyProperties(size_t count) {
	return (VkQueueFamilyProperties*)calloc(count, sizeof(VkQueueFamilyProperties));
}

static VkPhysicalDevice* allocPhysicalDevices(size_t count) {
	return (VkPhysicalDevice*)calloc(count, sizeof(VkPhysicalDevice));
}

static VkDescriptorSetLayoutBinding* allocDescriptorSetLayoutBindings(size_t count) {
	return (VkDescriptorSetLayoutBinding*)calloc(count, sizeof(VkDescriptorSetLayoutBinding));
}

static VkDescriptorPoolSize* allocDescriptorPoolSizes(size_t count) {
	return (VkDescriptorPoolSize*)calloc(count, sizeof(VkDescriptorPoolSize));
}

static VkWriteDescriptorSet* allocWriteDescriptorSets(size_t count) {
	return (VkWriteDescriptorSet*)calloc(count, sizeof(VkWriteDescriptorSet));
}

static VkDescriptorSet* allocDescriptorSets(size_t count) {
	return (VkDescriptorSet*)calloc(count, sizeof(VkDescriptorSet));
}

static VkPipelineShaderStageCreateInfo* allocPipelineShaderStages(size_t count) {
	return (VkPipelineShaderStageCreateInfo*)calloc(count, sizeof(VkPipelineShaderStageCreateInfo));
}

static VkComputePipelineCreateInfo* allocComputePipelineCreateInfos(size_t count) {
	return (VkComputePipelineCreateInfo*)calloc(count, sizeof(VkComputePipelineCreateInfo));
}

static VkCommandBuffer* allocCommandBuffers(size_t count) {
	return (VkCommandBuffer*)calloc(count, sizeof(VkCommandBuffer));
}

static void freePtr(void* p) {
	free(p);
}
*/
import "C"

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unsafe"

	"forza-painter-geometrize-go/internal/model"
)

// The Vulkan backend is intentionally native-cgo now: no vulkan-go bindings,
// no generated compatibility shim. We bind the small compute-only surface we
// actually use against the system SDK shipped in C:\VulkanSDK\1.4.350.0.

var (
	vkMaxTimeout C.uint64_t     = C.uint64_t(^uint64(0))
	vkWholeSize  C.VkDeviceSize = C.VK_WHOLE_SIZE
)

type vkHandle = C.VkInstance
type vkPhysicalDevice = C.VkPhysicalDevice
type vkDevice = C.VkDevice
type vkQueue = C.VkQueue
type vkCommandPool = C.VkCommandPool
type vkBuffer = C.VkBuffer
type vkDeviceMemory = C.VkDeviceMemory
type vkDescriptorSetLayout = C.VkDescriptorSetLayout
type vkDescriptorPool = C.VkDescriptorPool
type vkDescriptorSet = C.VkDescriptorSet
type vkPipeline = C.VkPipeline
type vkPipelineLayout = C.VkPipelineLayout
type vkFence = C.VkFence
type vkSemaphore = C.VkSemaphore
type vkCommandBuffer = C.VkCommandBuffer

type vkHandleInt = C.uintptr_t

type vkEvalSlot struct {
	seq  uint64
	busy bool
}

type vkGridSlot struct {
	seq  uint64
	busy bool
}

var spvEvalV3 []byte
var spvEvalV4 []byte
var spvApply []byte
var spvErrorGrid []byte

var vulkanInitOnce sync.Once
var vulkanInitErr error
var shaderSearchRoots []string

func init() {
	if exe, err := os.Executable(); err == nil {
		shaderSearchRoots = append(shaderSearchRoots, filepath.Join(filepath.Dir(exe), "shaders"))
	}
	shaderSearchRoots = append(shaderSearchRoots, "shaders")

	load := func(name string) []byte {
		for _, root := range shaderSearchRoots {
			p := filepath.Join(root, name)
			b, err := os.ReadFile(p)
			if err == nil {
				return b
			}
		}
		return nil
	}

	spvEvalV3 = load("eval_v3.comp.spv")
	spvEvalV4 = load("eval_v4.comp.spv")
	spvApply = load("apply.comp.spv")
	spvErrorGrid = load("error_grid.comp.spv")
}

type vulkanBackend struct {
	instance       vkHandle
	physicalDevice vkPhysicalDevice
	device         vkDevice
	queue          vkQueue
	queueFamilyIdx uint32

	commandPool vkCommandPool

	targetBuf  vkBuffer
	targetMem  vkDeviceMemory
	currentBuf vkBuffer
	currentMem vkDeviceMemory
	maskBuf    vkBuffer
	maskMem    vkDeviceMemory

	stagingBuf    vkBuffer
	stagingMem    vkDeviceMemory
	stagingSize   int
	stagingMapped unsafe.Pointer

	candBufs     [ringSize]vkBuffer
	candMems     [ringSize]vkDeviceMemory
	candMapped   [ringSize]unsafe.Pointer
	resultBufs   [ringSize]vkBuffer
	resultMems   [ringSize]vkDeviceMemory
	resultMapped [ringSize]unsafe.Pointer
	gridBufs     [ringSize]vkBuffer
	gridMems     [ringSize]vkDeviceMemory
	gridMapped   [ringSize]unsafe.Pointer

	dsLayout vkDescriptorSetLayout
	dsPool   vkDescriptorPool
	ds       [ringSize]vkDescriptorSet
	evalDs   [ringSize]vkDescriptorSet
	applyDs  [ringSize]vkDescriptorSet
	gridDs   [ringSize]vkDescriptorSet

	evalPipeV3   vkPipeline
	evalPipeV4   vkPipeline
	applyPipe    vkPipeline
	gridPipe     vkPipeline
	evalLayoutV3 vkPipelineLayout
	evalLayoutV4 vkPipelineLayout
	applyLayout  vkPipelineLayout
	gridLayout   vkPipelineLayout

	cmdBufs [ringSize]vkCommandBuffer
	fences  [ringSize]vkFence

	gridCmdBufs [ringSize]vkCommandBuffer
	gridFences  [ringSize]vkFence

	applyCmdBufs  [ringSize]vkCommandBuffer
	applyFences   [ringSize]vkFence
	applySlots    [ringSize]vkEvalSlot
	nextApplySlot int
	applySeq      uint64

	transferCmdBuf vkCommandBuffer
	transferFence  vkFence

	submitMu sync.Mutex

	useWorkGroupEval bool
	sampleStep       int

	evalSlots    [ringSize]vkEvalSlot
	nextEvalSlot int
	evalSeq      uint64

	gridSlots    [ringSize]vkGridSlot
	nextGridSlot int
	gridSeq      uint64

	width         int
	height        int
	pixelCount    int
	maxCandidates int
	gridW         int
	gridH         int

	closed bool
}

func newVulkanBackend(target, current []float32, maskData []uint8, width, height, maxCandidates, gridSize int) (Backend, error) {
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid dimensions: %dx%d", width, height)
	}
	if maxCandidates < 1 {
		return nil, fmt.Errorf("maxCandidates must be > 0")
	}
	if len(target) != len(current) || len(target) != width*height*4 {
		return nil, fmt.Errorf("buffer size mismatch")
	}
	if len(maskData) != width*height {
		return nil, fmt.Errorf("mask size mismatch: got %d, expected %d", len(maskData), width*height)
	}

	vulkanInitOnce.Do(func() {
		if err := ensureVulkanLoaded(); err != nil {
			vulkanInitErr = err
		}
	})
	if vulkanInitErr != nil {
		return nil, vulkanInitErr
	}

	inst, physDev, qIdx, err := vkPickDevice()
	if err != nil {
		return nil, err
	}
	dev, queue, err := vkMakeDevice(physDev, qIdx)
	if err != nil {
		C.rawVkDestroyInstance(inst, nil)
		return nil, err
	}
	cmdPool, err := vkMakeCommandPool(dev, qIdx)
	if err != nil {
		C.rawVkDestroyDevice(dev, nil)
		C.rawVkDestroyInstance(inst, nil)
		return nil, err
	}

	gridW := gridSize
	gridH := gridSize
	if width < gridW {
		gridW = width
	}
	if height < gridH {
		gridH = height
	}
	if gridW < 1 {
		gridW = 1
	}
	if gridH < 1 {
		gridH = 1
	}

	v := &vulkanBackend{
		instance:       inst,
		physicalDevice: physDev,
		device:         dev,
		queue:          queue,
		queueFamilyIdx: qIdx,
		commandPool:    cmdPool,
		width:          width,
		height:         height,
		pixelCount:     width * height,
		maxCandidates:  maxCandidates,
		gridW:          gridW,
		gridH:          gridH,
		sampleStep:     1,
	}

	// Pack the per-pixel opaque mask into 1-bit-per-pixel words.
	// shaders read it via bit-test, which is dramatically cheaper than
	// fetching a whole float per pixel. Each uint32 stores 32 pixels:
	// pixel p is opaque iff (mask[p>>5] >> (p & 31)) & 1.
	maskPackedLen := (v.pixelCount + 31) / 32
	maskPacked := make([]uint32, maskPackedLen)
	for i, m := range maskData {
		if m != 0 {
			maskPacked[i>>5] |= 1 << (uint(i) & 31)
		}
	}

	targetSize := len(target) * 4
	currentSize := len(current) * 4
	maskSize := maskPackedLen * 4
	candSize := maxCandidates * 7 * 4
	resultSize := maxCandidates * 4 * 4
	gridBufSize := gridW * gridH * 4

	cleanup := func() {
		v.destroy()
	}

	if err := v.makeDeviceBuffer(targetSize, C.VK_BUFFER_USAGE_TRANSFER_DST_BIT|C.VK_BUFFER_USAGE_STORAGE_BUFFER_BIT, &v.targetBuf, &v.targetMem); err != nil {
		cleanup()
		return nil, fmt.Errorf("target buffer: %w", err)
	}
	if err := v.makeDeviceBuffer(currentSize, C.VK_BUFFER_USAGE_TRANSFER_SRC_BIT|C.VK_BUFFER_USAGE_TRANSFER_DST_BIT|C.VK_BUFFER_USAGE_STORAGE_BUFFER_BIT, &v.currentBuf, &v.currentMem); err != nil {
		cleanup()
		return nil, fmt.Errorf("current buffer: %w", err)
	}
	if err := v.makeDeviceBuffer(maskSize, C.VK_BUFFER_USAGE_TRANSFER_DST_BIT|C.VK_BUFFER_USAGE_STORAGE_BUFFER_BIT, &v.maskBuf, &v.maskMem); err != nil {
		cleanup()
		return nil, fmt.Errorf("mask buffer: %w", err)
	}
	maxStaging := targetSize
	if currentSize > maxStaging {
		maxStaging = currentSize
	}
	if err := v.makeStagingBuffer(maxStaging); err != nil {
		cleanup()
		return nil, fmt.Errorf("staging buffer: %w", err)
	}
	if err := v.allocCommandBuffers(); err != nil {
		cleanup()
		return nil, fmt.Errorf("command buffers: %w", err)
	}
	if err := v.allocGridResources(); err != nil {
		cleanup()
		return nil, fmt.Errorf("grid resources: %w", err)
	}
	if err := v.allocTransferResources(); err != nil {
		cleanup()
		return nil, fmt.Errorf("transfer resources: %w", err)
	}
	if err := v.allocApplyResources(); err != nil {
		cleanup()
		return nil, fmt.Errorf("apply resources: %w", err)
	}
	if err := v.makeFences(); err != nil {
		cleanup()
		return nil, fmt.Errorf("fences: %w", err)
	}
	if err := v.upload(target, v.targetBuf, targetSize); err != nil {
		cleanup()
		return nil, fmt.Errorf("upload target: %w", err)
	}
	if err := v.upload(current, v.currentBuf, currentSize); err != nil {
		cleanup()
		return nil, fmt.Errorf("upload current: %w", err)
	}
	if maskSize > 0 {
		if err := v.uploadRaw(unsafe.Pointer(&maskPacked[0]), v.maskBuf, maskSize); err != nil {
			cleanup()
			return nil, fmt.Errorf("upload mask: %w", err)
		}
	}

	for i := 0; i < ringSize; i++ {
		if err := v.makeHostBuffer(candSize, &v.candBufs[i], &v.candMems[i], &v.candMapped[i]); err != nil {
			cleanup()
			return nil, fmt.Errorf("cand buffer %d: %w", i, err)
		}
		if err := v.makeHostBuffer(resultSize, &v.resultBufs[i], &v.resultMems[i], &v.resultMapped[i]); err != nil {
			cleanup()
			return nil, fmt.Errorf("result buffer %d: %w", i, err)
		}
		if err := v.makeHostBuffer(gridBufSize, &v.gridBufs[i], &v.gridMems[i], &v.gridMapped[i]); err != nil {
			cleanup()
			return nil, fmt.Errorf("grid buffer %d: %w", i, err)
		}
	}

	if err := v.makeDescriptorSetLayout(); err != nil {
		cleanup()
		return nil, fmt.Errorf("descriptor set layout: %w", err)
	}
	if err := v.allocCommandBuffers(); err != nil {
		cleanup()
		return nil, fmt.Errorf("command buffers: %w", err)
	}
	if err := v.allocTransferResources(); err != nil {
		cleanup()
		return nil, fmt.Errorf("transfer resources: %w", err)
	}
	if err := v.allocApplyResources(); err != nil {
		cleanup()
		return nil, fmt.Errorf("apply resources: %w", err)
	}
	if err := v.makePipelines(); err != nil {
		cleanup()
		return nil, fmt.Errorf("pipelines: %w", err)
	}
	if err := v.makeDescriptorPool(); err != nil {
		cleanup()
		return nil, fmt.Errorf("descriptor pool: %w", err)
	}
	if err := v.allocDescriptorSets(); err != nil {
		cleanup()
		return nil, fmt.Errorf("descriptor sets: %w", err)
	}
	v.writeDescriptorSets()
	return v, nil
}

func (v *vulkanBackend) destroy() {
	if v.device == nil {
		return
	}
	C.rawVkDeviceWaitIdle(v.device)
	if v.transferFence != nil {
		C.rawVkDestroyFence(v.device, v.transferFence, nil)
		v.transferFence = nil
	}
	if v.transferCmdBuf != nil {
		C.rawVkFreeCommandBuffers(v.device, v.commandPool, 1, &v.transferCmdBuf)
		v.transferCmdBuf = nil
	}
	for i := 0; i < ringSize; i++ {
		if v.applyFences[i] != nil {
			C.rawVkDestroyFence(v.device, v.applyFences[i], nil)
			v.applyFences[i] = nil
		}
	}
	if v.applyCmdBufs[0] != nil {
		C.rawVkFreeCommandBuffers(v.device, v.commandPool, C.uint32_t(len(v.applyCmdBufs)), &v.applyCmdBufs[0])
		v.applyCmdBufs = [ringSize]vkCommandBuffer{}
	}
	for i := 0; i < ringSize; i++ {
		if v.fences[i] != nil {
			C.rawVkDestroyFence(v.device, v.fences[i], nil)
			v.fences[i] = nil
		}
	}
	if v.cmdBufs[0] != nil {
		C.rawVkFreeCommandBuffers(v.device, v.commandPool, C.uint32_t(len(v.cmdBufs)), &v.cmdBufs[0])
		v.cmdBufs = [ringSize]vkCommandBuffer{}
	}
	for i := 0; i < ringSize; i++ {
		if v.gridFences[i] != nil {
			C.rawVkDestroyFence(v.device, v.gridFences[i], nil)
			v.gridFences[i] = nil
		}
	}
	if v.gridCmdBufs[0] != nil {
		C.rawVkFreeCommandBuffers(v.device, v.commandPool, C.uint32_t(len(v.gridCmdBufs)), &v.gridCmdBufs[0])
		v.gridCmdBufs = [ringSize]vkCommandBuffer{}
	}
	if v.dsPool != nil {
		C.rawVkDestroyDescriptorPool(v.device, v.dsPool, nil)
	}
	if v.evalPipeV3 != nil {
		C.rawVkDestroyPipeline(v.device, v.evalPipeV3, nil)
	}
	if v.evalPipeV4 != nil {
		C.rawVkDestroyPipeline(v.device, v.evalPipeV4, nil)
	}
	if v.applyPipe != nil {
		C.rawVkDestroyPipeline(v.device, v.applyPipe, nil)
	}
	if v.gridPipe != nil {
		C.rawVkDestroyPipeline(v.device, v.gridPipe, nil)
	}
	if v.evalLayoutV3 != nil {
		C.rawVkDestroyPipelineLayout(v.device, v.evalLayoutV3, nil)
	}
	if v.evalLayoutV4 != nil {
		C.rawVkDestroyPipelineLayout(v.device, v.evalLayoutV4, nil)
	}
	if v.applyLayout != nil {
		C.rawVkDestroyPipelineLayout(v.device, v.applyLayout, nil)
	}
	if v.gridLayout != nil {
		C.rawVkDestroyPipelineLayout(v.device, v.gridLayout, nil)
	}
	if v.dsLayout != nil {
		C.rawVkDestroyDescriptorSetLayout(v.device, v.dsLayout, nil)
	}
	for i := 0; i < ringSize; i++ {
		if v.candBufs[i] != nil {
			C.rawVkDestroyBuffer(v.device, v.candBufs[i], nil)
			C.rawVkFreeMemory(v.device, v.candMems[i], nil)
		}
		if v.resultBufs[i] != nil {
			C.rawVkDestroyBuffer(v.device, v.resultBufs[i], nil)
			C.rawVkFreeMemory(v.device, v.resultMems[i], nil)
		}
		if v.gridBufs[i] != nil {
			C.rawVkDestroyBuffer(v.device, v.gridBufs[i], nil)
			C.rawVkFreeMemory(v.device, v.gridMems[i], nil)
		}
	}
	if v.stagingBuf != nil {
		C.rawVkDestroyBuffer(v.device, v.stagingBuf, nil)
		C.rawVkFreeMemory(v.device, v.stagingMem, nil)
	}
	if v.maskBuf != nil {
		C.rawVkDestroyBuffer(v.device, v.maskBuf, nil)
		C.rawVkFreeMemory(v.device, v.maskMem, nil)
	}
	if v.currentBuf != nil {
		C.rawVkDestroyBuffer(v.device, v.currentBuf, nil)
		C.rawVkFreeMemory(v.device, v.currentMem, nil)
	}
	if v.targetBuf != nil {
		C.rawVkDestroyBuffer(v.device, v.targetBuf, nil)
		C.rawVkFreeMemory(v.device, v.targetMem, nil)
	}
	if v.commandPool != nil {
		C.rawVkDestroyCommandPool(v.device, v.commandPool, nil)
	}
	C.rawVkDestroyDevice(v.device, nil)
	C.rawVkDestroyInstance(v.instance, nil)
}

func (v *vulkanBackend) Close() error {
	if v.closed {
		return nil
	}
	v.closed = true
	_ = v.Flush()
	v.destroy()
	return nil
}

func (v *vulkanBackend) Flush() error {
	if v.device == nil {
		return nil
	}
	C.rawVkDeviceWaitIdle(v.device)
	for i := 0; i < ringSize; i++ {
		if v.applySlots[i].busy {
			v.applySlots[i].busy = false
			v.applySlots[i].seq = 0
		}
	}
	for i := 0; i < ringSize; i++ {
		if v.evalSlots[i].busy {
			v.evalSlots[i].busy = false
			v.evalSlots[i].seq = 0
		}
		if v.gridSlots[i].busy {
			v.gridSlots[i].busy = false
			v.gridSlots[i].seq = 0
		}
		if v.applySlots[i].busy {
			v.applySlots[i].busy = false
			v.applySlots[i].seq = 0
		}
	}
	for i := 0; i < ringSize; i++ {
		if v.fences[i] != nil {
			fence := v.fences[i]
			C.rawVkResetFencesOne(v.device, fence)
		}
		if v.gridFences[i] != nil {
			fence := v.gridFences[i]
			C.rawVkResetFencesOne(v.device, fence)
		}
		if v.applyFences[i] != nil {
			fence := v.applyFences[i]
			C.rawVkResetFencesOne(v.device, fence)
		}
	}
	if v.transferFence != nil {
		C.rawVkResetFencesOne(v.device, v.transferFence)
	}
	return nil
}

func (v *vulkanBackend) SetUseWorkGroupEval(val bool) { v.useWorkGroupEval = val }
func (v *vulkanBackend) SetSampleStep(val int)        { v.sampleStep = val }

// SetErrorMetric is a no-op on Vulkan — SSIM is only supported via OpenCL.
func (v *vulkanBackend) SetErrorMetric(metric string) {}

// SetSsimWeight is a no-op on Vulkan — SSIM is only supported via OpenCL.
func (v *vulkanBackend) SetSsimWeight(w float32) {}

// SubmitSsimMap returns an error on Vulkan — SSIM is only supported via OpenCL.
func (v *vulkanBackend) SubmitSsimMap() error {
	return fmt.Errorf("SubmitSsimMap: not supported on Vulkan backend")
}

func (v *vulkanBackend) SubmitEval(cands []model.Candidate) (EvalTicket, error) {
	count := len(cands)
	if count == 0 {
		return EvalTicket{}, nil
	}
	if count > v.maxCandidates {
		return EvalTicket{}, fmt.Errorf("candidate count %d exceeds max %d", count, v.maxCandidates)
	}

	slot := v.nextEvalSlot
	v.nextEvalSlot = (v.nextEvalSlot + 1) % ringSize
	if v.evalSlots[slot].busy {
		t := EvalTicket{slot: slot, seq: v.evalSlots[slot].seq, count: 0, valid: true}
		if _, err := v.WaitEval(t); err != nil {
			return EvalTicket{}, err
		}
	}
	dst := unsafe.Slice((*float32)(v.candMapped[slot]), count*7)
	for i, c := range cands {
		base := i * 7
		dst[base+0] = c.X
		dst[base+1] = c.Y
		dst[base+2] = c.RX
		dst[base+3] = c.RY
		dst[base+4] = c.Theta
		dst[base+5] = c.A
		dst[base+6] = float32(c.ShapeType)
	}

	if err := v.resetFence(slot); err != nil {
		return EvalTicket{}, err
	}
	cmdBuf := v.cmdBufs[slot]
	if err := v.beginCommandBuffer(cmdBuf); err != nil {
		return EvalTicket{}, err
	}

	hostBarrier := C.VkBufferMemoryBarrier{
		sType:               C.VK_STRUCTURE_TYPE_BUFFER_MEMORY_BARRIER,
		pNext:               nil,
		srcAccessMask:       C.VK_ACCESS_HOST_WRITE_BIT,
		dstAccessMask:       C.VK_ACCESS_SHADER_READ_BIT,
		srcQueueFamilyIndex: C.VK_QUEUE_FAMILY_IGNORED,
		dstQueueFamilyIndex: C.VK_QUEUE_FAMILY_IGNORED,
		buffer:              v.candBufs[slot],
		offset:              0,
		size:                C.VK_WHOLE_SIZE,
	}
	C.rawVkCmdPipelineBarrier(cmdBuf, C.VK_PIPELINE_STAGE_HOST_BIT, C.VK_PIPELINE_STAGE_COMPUTE_SHADER_BIT, 0, 0, nil, 1, &hostBarrier, 0, nil)

	var pipeline vkPipeline
	var layout vkPipelineLayout
	var dispatchX, dispatchY uint32
	if v.useWorkGroupEval && v.evalPipeV4 != nil {
		pipeline = v.evalPipeV4
		layout = v.evalLayoutV4
		// eval_v4 uses gl_WorkGroupID.x as the candidate index, so we must
		// launch one work-group per candidate.
		dispatchX = uint32(count)
		dispatchY = 1
	} else {
		pipeline = v.evalPipeV3
		layout = v.evalLayoutV3
		dispatchX = uint32((count + 63) / 64)
		dispatchY = 1
	}

	C.rawVkCmdBindPipeline(cmdBuf, C.VK_PIPELINE_BIND_POINT_COMPUTE, pipeline)
	ds := v.evalDs[slot]
	C.rawVkCmdBindDescriptorSets(cmdBuf, C.VK_PIPELINE_BIND_POINT_COMPUTE, layout, 0, 1, &ds)
	pc := [4]int32{int32(v.width), int32(v.height), int32(v.sampleStep), int32(count)}
	C.rawVkCmdPushConstants(cmdBuf, layout, C.VK_SHADER_STAGE_COMPUTE_BIT, 0, C.uint32_t(unsafe.Sizeof(pc)), unsafe.Pointer(&pc[0]))
	C.rawVkCmdDispatch(cmdBuf, C.uint32_t(dispatchX), C.uint32_t(dispatchY), 1)

	resultBarrier := C.VkBufferMemoryBarrier{
		sType:               C.VK_STRUCTURE_TYPE_BUFFER_MEMORY_BARRIER,
		pNext:               nil,
		srcAccessMask:       C.VK_ACCESS_SHADER_WRITE_BIT,
		dstAccessMask:       C.VK_ACCESS_HOST_READ_BIT,
		srcQueueFamilyIndex: C.VK_QUEUE_FAMILY_IGNORED,
		dstQueueFamilyIndex: C.VK_QUEUE_FAMILY_IGNORED,
		buffer:              v.resultBufs[slot],
		offset:              0,
		size:                C.VK_WHOLE_SIZE,
	}
	C.rawVkCmdPipelineBarrier(cmdBuf, C.VK_PIPELINE_STAGE_COMPUTE_SHADER_BIT, C.VK_PIPELINE_STAGE_HOST_BIT, 0, 0, nil, 1, &resultBarrier, 0, nil)
	if err := v.endCommandBuffer(cmdBuf); err != nil {
		return EvalTicket{}, err
	}
	if err := v.submit("eval", cmdBuf, v.fences[slot]); err != nil {
		return EvalTicket{}, err
	}

	v.evalSeq++
	v.evalSlots[slot] = vkEvalSlot{seq: v.evalSeq, busy: true}
	return EvalTicket{slot: slot, seq: v.evalSeq, count: count, valid: true}, nil
}

func (v *vulkanBackend) WaitEval(t EvalTicket) ([]EvalResult, error) {
	if !t.valid {
		return nil, nil
	}
	s := &v.evalSlots[t.slot]
	if !s.busy || s.seq != t.seq {
		return nil, fmt.Errorf("WaitEval: stale or invalid ticket")
	}
	if err := v.waitFence(t.slot); err != nil {
		return nil, err
	}
	s.busy = false
	src := unsafe.Slice((*float32)(v.resultMapped[t.slot]), t.count*4)
	out := make([]EvalResult, t.count)
	for i := 0; i < t.count; i++ {
		out[i] = EvalResult{Score: src[i*4+0], R: src[i*4+1], G: src[i*4+2], B: src[i*4+3]}
	}
	return out, nil
}

func (v *vulkanBackend) Evaluate(cands []model.Candidate) ([]EvalResult, error) {
	t, err := v.SubmitEval(cands)
	if err != nil {
		return nil, err
	}
	return v.WaitEval(t)
}

func (v *vulkanBackend) SubmitApply(candidate model.Candidate) error {
	rx := candidate.RX
	ry := candidate.RY
	if rx < 1 {
		rx = 1
	}
	if ry < 1 {
		ry = 1
	}
	theta := float64(candidate.Theta) * (math.Pi / 180.0)
	cosT := math.Cos(theta)
	sinT := math.Sin(theta)
	rx2 := float64(rx) * float64(rx)
	ry2 := float64(ry) * float64(ry)
	cos2 := cosT * cosT
	sin2 := sinT * sinT
	ex := math.Sqrt(rx2*cos2 + ry2*sin2)
	ey := math.Sqrt(rx2*sin2 + ry2*cos2)

	xMin := int(math.Floor(float64(candidate.X) - ex - 1.0))
	xMax := int(math.Ceil(float64(candidate.X) + ex + 1.0))
	yMin := int(math.Floor(float64(candidate.Y) - ey - 1.0))
	yMax := int(math.Ceil(float64(candidate.Y) + ey + 1.0))
	if xMin < 0 {
		xMin = 0
	}
	if yMin < 0 {
		yMin = 0
	}
	if xMax > v.width-1 {
		xMax = v.width - 1
	}
	if yMax > v.height-1 {
		yMax = v.height - 1
	}
	if xMax < xMin || yMax < yMin {
		return nil
	}
	bw := xMax - xMin + 1
	bh := yMax - yMin + 1
	slot := v.nextApplySlot
	v.nextApplySlot = (v.nextApplySlot + 1) % ringSize
	if v.applySlots[slot].busy {
		if err := v.waitApplySlot(slot); err != nil {
			return err
		}
	}
	if err := v.resetApplyFence(slot); err != nil {
		return err
	}
	cmdBuf := v.applyCmdBufs[slot]
	if err := v.beginCommandBuffer(cmdBuf); err != nil {
		return err
	}
	if err := v.bindApplyArgs(cmdBuf, slot, xMin, yMin, xMax, yMax, candidate); err != nil {
		return err
	}
	dx := uint32((bw + 7) / 8)
	dy := uint32((bh + 7) / 8)
	C.rawVkCmdDispatch(cmdBuf, C.uint32_t(dx), C.uint32_t(dy), 1)
	curBarrier := C.VkBufferMemoryBarrier{
		sType:               C.VK_STRUCTURE_TYPE_BUFFER_MEMORY_BARRIER,
		pNext:               nil,
		srcAccessMask:       C.VK_ACCESS_SHADER_WRITE_BIT,
		dstAccessMask:       C.VK_ACCESS_SHADER_READ_BIT,
		srcQueueFamilyIndex: C.VK_QUEUE_FAMILY_IGNORED,
		dstQueueFamilyIndex: C.VK_QUEUE_FAMILY_IGNORED,
		buffer:              v.currentBuf,
		offset:              0,
		size:                C.VK_WHOLE_SIZE,
	}
	C.rawVkCmdPipelineBarrier(cmdBuf, C.VK_PIPELINE_STAGE_COMPUTE_SHADER_BIT, C.VK_PIPELINE_STAGE_COMPUTE_SHADER_BIT, 0, 0, nil, 1, &curBarrier, 0, nil)
	if err := v.endCommandBuffer(cmdBuf); err != nil {
		return err
	}
	if err := v.submit("apply", cmdBuf, v.applyFences[slot]); err != nil {
		return err
	}
	v.applySeq++
	v.applySlots[slot] = vkEvalSlot{seq: v.applySeq, busy: true}
	return nil
}

func (v *vulkanBackend) Apply(candidate model.Candidate) error {
	if err := v.SubmitApply(candidate); err != nil {
		return err
	}
	return v.Flush()
}

func (v *vulkanBackend) SubmitErrorGrid() (GridTicket, error) {
	slot := v.nextGridSlot
	v.nextGridSlot = (v.nextGridSlot + 1) % ringSize
	if v.gridSlots[slot].busy {
		t := GridTicket{slot: slot, seq: v.gridSlots[slot].seq, valid: true}
		if _, _, _, err := v.WaitErrorGrid(t); err != nil {
			return GridTicket{}, err
		}
	}
	if err := v.resetGridFence(slot); err != nil {
		return GridTicket{}, err
	}
	cmdBuf := v.gridCmdBufs[slot]
	if err := v.beginCommandBuffer(cmdBuf); err != nil {
		return GridTicket{}, err
	}
	if err := v.bindGridArgs(cmdBuf, slot); err != nil {
		return GridTicket{}, err
	}
	C.rawVkCmdDispatch(cmdBuf, C.uint32_t((v.gridW+7)/8), C.uint32_t((v.gridH+7)/8), 1)
	gridBarrier := C.VkBufferMemoryBarrier{
		sType:               C.VK_STRUCTURE_TYPE_BUFFER_MEMORY_BARRIER,
		pNext:               nil,
		srcAccessMask:       C.VK_ACCESS_SHADER_WRITE_BIT,
		dstAccessMask:       C.VK_ACCESS_HOST_READ_BIT,
		srcQueueFamilyIndex: C.VK_QUEUE_FAMILY_IGNORED,
		dstQueueFamilyIndex: C.VK_QUEUE_FAMILY_IGNORED,
		buffer:              v.gridBufs[slot],
		offset:              0,
		size:                C.VK_WHOLE_SIZE,
	}
	C.rawVkCmdPipelineBarrier(cmdBuf, C.VK_PIPELINE_STAGE_COMPUTE_SHADER_BIT, C.VK_PIPELINE_STAGE_HOST_BIT, 0, 0, nil, 1, &gridBarrier, 0, nil)
	if err := v.endCommandBuffer(cmdBuf); err != nil {
		return GridTicket{}, err
	}
	if err := v.submit("grid", cmdBuf, v.gridFences[slot]); err != nil {
		return GridTicket{}, err
	}
	v.gridSeq++
	v.gridSlots[slot] = vkGridSlot{seq: v.gridSeq, busy: true}
	return GridTicket{slot: slot, seq: v.gridSeq, valid: true}, nil
}

func (v *vulkanBackend) WaitErrorGrid(t GridTicket) ([]float32, int, int, error) {
	if !t.valid {
		return nil, 0, 0, fmt.Errorf("WaitErrorGrid: invalid ticket")
	}
	s := &v.gridSlots[t.slot]
	if !s.busy || s.seq != t.seq {
		return nil, 0, 0, fmt.Errorf("WaitErrorGrid: stale or invalid ticket")
	}
	if err := v.waitGridFence(t.slot); err != nil {
		return nil, 0, 0, err
	}
	s.busy = false
	n := v.gridW * v.gridH
	src := unsafe.Slice((*float32)(v.gridMapped[t.slot]), n)
	out := make([]float32, n)
	copy(out, src)
	return out, v.gridW, v.gridH, nil
}

func (v *vulkanBackend) ErrorGrid() ([]float32, int, int, error) {
	t, err := v.SubmitErrorGrid()
	if err != nil {
		return nil, 0, 0, err
	}
	return v.WaitErrorGrid(t)
}

func (v *vulkanBackend) ReadCurrent(dst []float32) error {
	if len(dst) != v.pixelCount*4 {
		return fmt.Errorf("destination length mismatch")
	}
	if err := v.download(v.currentBuf, len(dst)*4); err != nil {
		return err
	}
	src := unsafe.Slice((*float32)(v.stagingMapped), len(dst))
	copy(dst, src)
	return nil
}

func (v *vulkanBackend) GridDims() (int, int)  { return v.gridW, v.gridH }
func (v *vulkanBackend) ImageDims() (int, int) { return v.width, v.height }

func (v *vulkanBackend) ResetCurrentBuffer(current []float32) error {
	if len(current) != v.pixelCount*4 {
		return fmt.Errorf("current slice length mismatch")
	}
	return v.upload(current, v.currentBuf, len(current)*4)
}

func (v *vulkanBackend) submit(kind string, cmdBuf vkCommandBuffer, fence vkFence) error {
	v.submitMu.Lock()
	defer v.submitMu.Unlock()
	res := C.rawVkQueueSubmitOne(v.queue, cmdBuf, fence)
	if res != C.VK_SUCCESS {
		return fmt.Errorf("%s: vkQueueSubmit failed: %d", kind, int(res))
	}
	return nil
}

func (v *vulkanBackend) resetFence(slot int) error {
	fence := v.fences[slot]
	if res := C.rawVkResetFencesOne(v.device, fence); res != C.VK_SUCCESS {
		return fmt.Errorf("vkResetFences[%d]: %d", slot, int(res))
	}
	return nil
}

func (v *vulkanBackend) resetApplyFence(slot int) error {
	fence := v.applyFences[slot]
	if res := C.rawVkResetFencesOne(v.device, fence); res != C.VK_SUCCESS {
		return fmt.Errorf("vkResetFences[apply:%d]: %d", slot, int(res))
	}
	return nil
}

func (v *vulkanBackend) waitTransferFence() error {
	if v.transferFence == nil {
		return fmt.Errorf("transfer fence not initialized")
	}
	if res := C.rawVkWaitForFencesOne(v.device, v.transferFence, vkMaxTimeout); res != C.VK_SUCCESS {
		return fmt.Errorf("vkWaitForFences[transfer]: %d", int(res))
	}
	return nil
}

func (v *vulkanBackend) resetTransferFence() error {
	if v.transferFence == nil {
		return fmt.Errorf("transfer fence not initialized")
	}
	if res := C.rawVkResetFencesOne(v.device, v.transferFence); res != C.VK_SUCCESS {
		return fmt.Errorf("vkResetFences[transfer]: %d", int(res))
	}
	return nil
}

func (v *vulkanBackend) resetGridFence(slot int) error {
	fence := v.gridFences[slot]
	if res := C.rawVkResetFencesOne(v.device, fence); res != C.VK_SUCCESS {
		return fmt.Errorf("vkResetFences[grid:%d]: %d", slot, int(res))
	}
	return nil
}

func (v *vulkanBackend) waitGridFence(slot int) error {
	fence := v.gridFences[slot]
	if res := C.rawVkWaitForFencesOne(v.device, fence, vkMaxTimeout); res != C.VK_SUCCESS {
		return fmt.Errorf("vkWaitForFences[grid:%d]: %d", slot, int(res))
	}
	return nil
}

func (v *vulkanBackend) waitApplySlot(slot int) error {
	fence := v.applyFences[slot]
	if res := C.rawVkWaitForFencesOne(v.device, fence, vkMaxTimeout); res != C.VK_SUCCESS {
		return fmt.Errorf("vkWaitForFences[apply:%d]: %d", slot, int(res))
	}
	v.applySlots[slot].busy = false
	return nil
}

func (v *vulkanBackend) writeApplyDescriptor(slot int) {
	C.rawVkUpdateDescriptorSetStorageBuffer(v.device, v.applyDs[slot], 1, v.currentBuf)
	C.rawVkUpdateDescriptorSetStorageBuffer(v.device, v.applyDs[slot], 2, v.maskBuf)
}

func (v *vulkanBackend) waitFence(slot int) error {
	fence := v.fences[slot]
	if res := C.rawVkWaitForFencesOne(v.device, fence, vkMaxTimeout); res != C.VK_SUCCESS {
		return fmt.Errorf("vkWaitForFences[%d]: %d", slot, int(res))
	}
	return nil
}

func (v *vulkanBackend) upload(data []float32, dstBuf vkBuffer, size int) error {
	if size <= 0 {
		return nil
	}
	return v.uploadRaw(unsafe.Pointer(&data[0]), dstBuf, size)
}

func (v *vulkanBackend) uploadRaw(srcPtr unsafe.Pointer, dstBuf vkBuffer, size int) error {
	if size <= 0 {
		return nil
	}
	dstStaging := unsafe.Slice((*byte)(v.stagingMapped), size)
	srcBytes := unsafe.Slice((*byte)(srcPtr), size)
	copy(dstStaging, srcBytes)
	if err := v.resetTransferFence(); err != nil {
		return err
	}
	cmdBuf := v.transferCmdBuf
	if err := v.beginCommandBuffer(cmdBuf); err != nil {
		return err
	}
	copyRegion := C.VkBufferCopy{size: C.VkDeviceSize(size)}
	C.rawVkCmdCopyBuffer(cmdBuf, v.stagingBuf, dstBuf, 1, &copyRegion)
	dstBarrier := C.VkBufferMemoryBarrier{
		sType:               C.VK_STRUCTURE_TYPE_BUFFER_MEMORY_BARRIER,
		pNext:               nil,
		srcAccessMask:       C.VK_ACCESS_TRANSFER_WRITE_BIT,
		dstAccessMask:       C.VK_ACCESS_SHADER_READ_BIT,
		srcQueueFamilyIndex: C.VK_QUEUE_FAMILY_IGNORED,
		dstQueueFamilyIndex: C.VK_QUEUE_FAMILY_IGNORED,
		buffer:              dstBuf,
		offset:              0,
		size:                C.VK_WHOLE_SIZE,
	}
	C.rawVkCmdPipelineBarrier(cmdBuf, C.VK_PIPELINE_STAGE_TRANSFER_BIT, C.VK_PIPELINE_STAGE_COMPUTE_SHADER_BIT, 0, 0, nil, 1, &dstBarrier, 0, nil)
	if err := v.endCommandBuffer(cmdBuf); err != nil {
		return err
	}
	if err := v.submit("transfer upload", cmdBuf, v.transferFence); err != nil {
		return err
	}
	if res := C.rawVkWaitForFencesOne(v.device, v.transferFence, vkMaxTimeout); res != C.VK_SUCCESS {
		return fmt.Errorf("vkWaitForFences[transfer]: %d", int(res))
	}
	return nil
}

func (v *vulkanBackend) download(srcBuf vkBuffer, size int) error {
	if err := v.resetTransferFence(); err != nil {
		return err
	}
	cmdBuf := v.transferCmdBuf
	if err := v.beginCommandBuffer(cmdBuf); err != nil {
		return err
	}
	copyRegion := C.VkBufferCopy{size: C.VkDeviceSize(size)}
	C.rawVkCmdCopyBuffer(cmdBuf, srcBuf, v.stagingBuf, 1, &copyRegion)
	hostBarrier := C.VkBufferMemoryBarrier{
		sType:               C.VK_STRUCTURE_TYPE_BUFFER_MEMORY_BARRIER,
		pNext:               nil,
		srcAccessMask:       C.VK_ACCESS_TRANSFER_WRITE_BIT,
		dstAccessMask:       C.VK_ACCESS_HOST_READ_BIT,
		srcQueueFamilyIndex: C.VK_QUEUE_FAMILY_IGNORED,
		dstQueueFamilyIndex: C.VK_QUEUE_FAMILY_IGNORED,
		buffer:              v.stagingBuf,
		offset:              0,
		size:                C.VK_WHOLE_SIZE,
	}
	C.rawVkCmdPipelineBarrier(cmdBuf, C.VK_PIPELINE_STAGE_TRANSFER_BIT, C.VK_PIPELINE_STAGE_HOST_BIT, 0, 0, nil, 1, &hostBarrier, 0, nil)
	if err := v.endCommandBuffer(cmdBuf); err != nil {
		return err
	}
	if err := v.submit("transfer download", cmdBuf, v.transferFence); err != nil {
		return err
	}
	if res := C.rawVkWaitForFencesOne(v.device, v.transferFence, vkMaxTimeout); res != C.VK_SUCCESS {
		return fmt.Errorf("vkWaitForFences[transfer]: %d", int(res))
	}
	return nil
}

func (v *vulkanBackend) makeDeviceBuffer(size int, usage C.VkBufferUsageFlags, buf *vkBuffer, mem *vkDeviceMemory) error {
	bci := C.VkBufferCreateInfo{
		sType:                 C.VK_STRUCTURE_TYPE_BUFFER_CREATE_INFO,
		pNext:                 nil,
		flags:                 0,
		size:                  C.VkDeviceSize(size),
		usage:                 usage,
		sharingMode:           C.VK_SHARING_MODE_EXCLUSIVE,
		queueFamilyIndexCount: 0,
		pQueueFamilyIndices:   nil,
	}
	if res := C.rawVkCreateBuffer(v.device, &bci, nil, buf); res != C.VK_SUCCESS {
		return fmt.Errorf("vkCreateBuffer: %d", int(res))
	}
	var req C.VkMemoryRequirements
	C.rawVkGetBufferMemoryRequirements(v.device, *buf, &req)
	memType, err := v.findMemoryType(uint32(req.memoryTypeBits), C.VK_MEMORY_PROPERTY_DEVICE_LOCAL_BIT)
	if err != nil {
		C.rawVkDestroyBuffer(v.device, *buf, nil)
		return err
	}
	mai := C.VkMemoryAllocateInfo{
		sType:           C.VK_STRUCTURE_TYPE_MEMORY_ALLOCATE_INFO,
		pNext:           nil,
		allocationSize:  req.size,
		memoryTypeIndex: C.uint32_t(memType),
	}
	if res := C.rawVkAllocateMemory(v.device, &mai, nil, mem); res != C.VK_SUCCESS {
		C.rawVkDestroyBuffer(v.device, *buf, nil)
		return fmt.Errorf("vkAllocateMemory: %d", int(res))
	}
	if res := C.rawVkBindBufferMemory(v.device, *buf, *mem, 0); res != C.VK_SUCCESS {
		C.rawVkDestroyBuffer(v.device, *buf, nil)
		C.rawVkFreeMemory(v.device, *mem, nil)
		return fmt.Errorf("vkBindBufferMemory: %d", int(res))
	}
	return nil
}

func (v *vulkanBackend) makeStagingBuffer(size int) error {
	if size <= 0 {
		size = 4
	}
	v.stagingSize = size
	bci := C.VkBufferCreateInfo{
		sType:       C.VK_STRUCTURE_TYPE_BUFFER_CREATE_INFO,
		pNext:       nil,
		flags:       0,
		size:        C.VkDeviceSize(size),
		usage:       C.VK_BUFFER_USAGE_TRANSFER_SRC_BIT | C.VK_BUFFER_USAGE_TRANSFER_DST_BIT,
		sharingMode: C.VK_SHARING_MODE_EXCLUSIVE,
	}
	if res := C.rawVkCreateBuffer(v.device, &bci, nil, &v.stagingBuf); res != C.VK_SUCCESS {
		return fmt.Errorf("vkCreateBuffer(staging): %d", int(res))
	}
	var req C.VkMemoryRequirements
	C.rawVkGetBufferMemoryRequirements(v.device, v.stagingBuf, &req)
	memType, err := v.findMemoryType(uint32(req.memoryTypeBits), C.VK_MEMORY_PROPERTY_HOST_VISIBLE_BIT|C.VK_MEMORY_PROPERTY_HOST_COHERENT_BIT)
	if err != nil {
		C.rawVkDestroyBuffer(v.device, v.stagingBuf, nil)
		return err
	}
	mai := C.VkMemoryAllocateInfo{
		sType:           C.VK_STRUCTURE_TYPE_MEMORY_ALLOCATE_INFO,
		pNext:           nil,
		allocationSize:  req.size,
		memoryTypeIndex: C.uint32_t(memType),
	}
	if res := C.rawVkAllocateMemory(v.device, &mai, nil, &v.stagingMem); res != C.VK_SUCCESS {
		C.rawVkDestroyBuffer(v.device, v.stagingBuf, nil)
		return fmt.Errorf("vkAllocateMemory(staging): %d", int(res))
	}
	if res := C.rawVkBindBufferMemory(v.device, v.stagingBuf, v.stagingMem, 0); res != C.VK_SUCCESS {
		C.rawVkDestroyBuffer(v.device, v.stagingBuf, nil)
		C.rawVkFreeMemory(v.device, v.stagingMem, nil)
		return fmt.Errorf("vkBindBufferMemory(staging): %d", int(res))
	}
	var mapped unsafe.Pointer
	if res := C.rawVkMapMemory(v.device, v.stagingMem, 0, C.VK_WHOLE_SIZE, 0, (*unsafe.Pointer)(unsafe.Pointer(&mapped))); res != C.VK_SUCCESS {
		C.rawVkDestroyBuffer(v.device, v.stagingBuf, nil)
		C.rawVkFreeMemory(v.device, v.stagingMem, nil)
		return fmt.Errorf("vkMapMemory(staging): %d", int(res))
	}
	v.stagingMapped = mapped
	return nil
}

func (v *vulkanBackend) makeHostBuffer(size int, buf *vkBuffer, mem *vkDeviceMemory, mapped *unsafe.Pointer) error {
	bci := C.VkBufferCreateInfo{
		sType:       C.VK_STRUCTURE_TYPE_BUFFER_CREATE_INFO,
		pNext:       nil,
		flags:       0,
		size:        C.VkDeviceSize(size),
		usage:       C.VK_BUFFER_USAGE_STORAGE_BUFFER_BIT,
		sharingMode: C.VK_SHARING_MODE_EXCLUSIVE,
	}
	if res := C.rawVkCreateBuffer(v.device, &bci, nil, buf); res != C.VK_SUCCESS {
		return fmt.Errorf("vkCreateBuffer(host): %d", int(res))
	}
	var req C.VkMemoryRequirements
	C.rawVkGetBufferMemoryRequirements(v.device, *buf, &req)
	memType, err := v.findMemoryType(uint32(req.memoryTypeBits), C.VK_MEMORY_PROPERTY_HOST_VISIBLE_BIT|C.VK_MEMORY_PROPERTY_HOST_COHERENT_BIT)
	if err != nil {
		C.rawVkDestroyBuffer(v.device, *buf, nil)
		return err
	}
	mai := C.VkMemoryAllocateInfo{
		sType:           C.VK_STRUCTURE_TYPE_MEMORY_ALLOCATE_INFO,
		pNext:           nil,
		allocationSize:  req.size,
		memoryTypeIndex: C.uint32_t(memType),
	}
	if res := C.rawVkAllocateMemory(v.device, &mai, nil, mem); res != C.VK_SUCCESS {
		C.rawVkDestroyBuffer(v.device, *buf, nil)
		return fmt.Errorf("vkAllocateMemory(host): %d", int(res))
	}
	if res := C.rawVkBindBufferMemory(v.device, *buf, *mem, 0); res != C.VK_SUCCESS {
		C.rawVkDestroyBuffer(v.device, *buf, nil)
		C.rawVkFreeMemory(v.device, *mem, nil)
		return fmt.Errorf("vkBindBufferMemory(host): %d", int(res))
	}
	if res := C.rawVkMapMemory(v.device, *mem, 0, C.VK_WHOLE_SIZE, 0, (*unsafe.Pointer)(unsafe.Pointer(mapped))); res != C.VK_SUCCESS {
		C.rawVkDestroyBuffer(v.device, *buf, nil)
		C.rawVkFreeMemory(v.device, *mem, nil)
		return fmt.Errorf("vkMapMemory(host): %d", int(res))
	}
	return nil
}

func (v *vulkanBackend) findMemoryType(typeFilter uint32, properties C.VkMemoryPropertyFlags) (uint32, error) {
	var memProps C.VkPhysicalDeviceMemoryProperties
	C.rawVkGetPhysicalDeviceMemoryProperties(v.physicalDevice, &memProps)
	for i := C.uint32_t(0); i < memProps.memoryTypeCount; i++ {
		if typeFilter&(1<<i) != 0 && memProps.memoryTypes[i].propertyFlags&properties == properties {
			return uint32(i), nil
		}
	}
	return 0, fmt.Errorf("no suitable memory type found (filter=0x%x props=0x%x)", typeFilter, uint32(properties))
}

func (v *vulkanBackend) makeDescriptorSetLayout() error {
	if res := C.rawVkCreateDescriptorSetLayoutSimple(v.device, &v.dsLayout); res != C.VK_SUCCESS {
		return fmt.Errorf("vkCreateDescriptorSetLayout: %d", int(res))
	}
	return nil
}

func (v *vulkanBackend) makeDescriptorPool() error {
	if res := C.rawVkCreateDescriptorPoolSized(v.device, 64, 320, &v.dsPool); res != C.VK_SUCCESS {
		return fmt.Errorf("vkCreateDescriptorPool: %d", int(res))
	}
	return nil
}

func (v *vulkanBackend) allocDescriptorSets() error {
	if res := C.rawVkAllocateDescriptorSetsSimple(v.device, v.dsPool, v.dsLayout, C.uint32_t(ringSize), &v.evalDs[0]); res != C.VK_SUCCESS {
		return fmt.Errorf("vkAllocateDescriptorSets(eval): %d", int(res))
	}
	if res := C.rawVkAllocateDescriptorSetsSimple(v.device, v.dsPool, v.dsLayout, C.uint32_t(ringSize), &v.applyDs[0]); res != C.VK_SUCCESS {
		return fmt.Errorf("vkAllocateDescriptorSets(apply): %d", int(res))
	}
	if res := C.rawVkAllocateDescriptorSetsSimple(v.device, v.dsPool, v.dsLayout, C.uint32_t(ringSize), &v.gridDs[0]); res != C.VK_SUCCESS {
		return fmt.Errorf("vkAllocateDescriptorSets(grid): %d", int(res))
	}
	return nil
}

func (v *vulkanBackend) writeDescriptorSets() {
	for i := 0; i < ringSize; i++ {
		C.rawVkUpdateDescriptorSetStorageBuffer(v.device, v.evalDs[i], 0, v.targetBuf)
		C.rawVkUpdateDescriptorSetStorageBuffer(v.device, v.evalDs[i], 1, v.currentBuf)
		C.rawVkUpdateDescriptorSetStorageBuffer(v.device, v.evalDs[i], 2, v.maskBuf)
		C.rawVkUpdateDescriptorSetStorageBuffer(v.device, v.evalDs[i], 3, v.candBufs[i])
		C.rawVkUpdateDescriptorSetStorageBuffer(v.device, v.evalDs[i], 4, v.resultBufs[i])
		C.rawVkUpdateDescriptorSetStorageBuffer(v.device, v.applyDs[i], 1, v.currentBuf)
		C.rawVkUpdateDescriptorSetStorageBuffer(v.device, v.applyDs[i], 2, v.maskBuf)
		C.rawVkUpdateDescriptorSetStorageBuffer(v.device, v.gridDs[i], 0, v.targetBuf)
		C.rawVkUpdateDescriptorSetStorageBuffer(v.device, v.gridDs[i], 1, v.currentBuf)
		C.rawVkUpdateDescriptorSetStorageBuffer(v.device, v.gridDs[i], 2, v.maskBuf)
		C.rawVkUpdateDescriptorSetStorageBuffer(v.device, v.gridDs[i], 4, v.gridBufs[i])
	}
}

func (v *vulkanBackend) makePipelines() error {
	if err := v.makeComputePipeline(&v.evalPipeV3, &v.evalLayoutV3, "eval_v3", spvEvalV3); err != nil {
		return err
	}
	if err := v.makeComputePipeline(&v.evalPipeV4, &v.evalLayoutV4, "eval_v4", spvEvalV4); err != nil {
		return err
	}
	if err := v.makeComputePipeline(&v.applyPipe, &v.applyLayout, "apply", spvApply); err != nil {
		return err
	}
	if err := v.makeComputePipeline(&v.gridPipe, &v.gridLayout, "error_grid", spvErrorGrid); err != nil {
		return err
	}
	return nil
}

func vkPickDevice() (vkHandle, vkPhysicalDevice, uint32, error) {
	var inst vkHandle
	appName := C.CString("ForzaPainterGeometrize")
	engineName := C.CString("forza-painter-geometrize-go")
	defer C.free(unsafe.Pointer(appName))
	defer C.free(unsafe.Pointer(engineName))
	if res := C.rawVkCreateInstanceSimple(appName, engineName, &inst); res != C.VK_SUCCESS {
		return inst, nil, 0, fmt.Errorf("vkCreateInstance: %d", int(res))
	}
	var physCount C.uint32_t
	if res := C.rawVkEnumeratePhysicalDevices(inst, &physCount, nil); res != C.VK_SUCCESS || physCount == 0 {
		C.rawVkDestroyInstance(inst, nil)
		return inst, nil, 0, fmt.Errorf("no Vulkan physical devices found")
	}
	phys := C.allocPhysicalDevices(C.size_t(physCount))
	if phys == nil {
		C.rawVkDestroyInstance(inst, nil)
		return inst, nil, 0, fmt.Errorf("allocPhysicalDevices failed")
	}
	defer C.freePtr(unsafe.Pointer(phys))
	if res := C.rawVkEnumeratePhysicalDevices(inst, &physCount, phys); res != C.VK_SUCCESS {
		C.rawVkDestroyInstance(inst, nil)
		return inst, nil, 0, fmt.Errorf("vkEnumeratePhysicalDevices: %d", int(res))
	}
	physSlice := unsafe.Slice((*vkPhysicalDevice)(unsafe.Pointer(phys)), int(physCount))
	var best vkPhysicalDevice
	var bestQI uint32
	var bestScore int64 = -1
	for _, pd := range physSlice {
		score, qi := scoreVkDevice(pd)
		if score > bestScore {
			bestScore = score
			best = pd
			bestQI = qi
		}
	}
	if best == nil {
		C.rawVkDestroyInstance(inst, nil)
		return inst, nil, 0, fmt.Errorf("no suitable Vulkan device found")
	}
	vkLogSelectedDevice(best)
	return inst, best, bestQI, nil
}

func vkLogSelectedDevice(pd vkPhysicalDevice) {
	var props C.VkPhysicalDeviceProperties
	C.rawVkGetPhysicalDeviceProperties(pd, &props)

	var memProps C.VkPhysicalDeviceMemoryProperties
	C.rawVkGetPhysicalDeviceMemoryProperties(pd, &memProps)
	var vramMB int64
	for i := C.uint32_t(0); i < memProps.memoryHeapCount; i++ {
		if memProps.memoryHeaps[i].flags&C.VK_MEMORY_HEAP_DEVICE_LOCAL_BIT != 0 {
			vramMB = int64(memProps.memoryHeaps[i].size / (1024 * 1024))
			break
		}
	}

	name := C.GoString(&props.deviceName[0])
	vendor := vkVendorName(uint32(props.vendorID))
	isGPU := props.deviceType == C.VK_PHYSICAL_DEVICE_TYPE_DISCRETE_GPU ||
		props.deviceType == C.VK_PHYSICAL_DEVICE_TYPE_INTEGRATED_GPU ||
		props.deviceType == C.VK_PHYSICAL_DEVICE_TYPE_VIRTUAL_GPU
	isDiscrete := props.deviceType == C.VK_PHYSICAL_DEVICE_TYPE_DISCRETE_GPU
	// Vulkan does not expose OpenCL-style compute units directly; keep the
	// log shape consistent and mark the field as not available.
	computeUnits := "n/a"

	fmt.Printf("Vulkan: Selected device %q (Vendor: %q, GPU: %v, Discrete: %v, VRAM: %dMB, Compute Units: %s)\n",
		name, vendor, isGPU, isDiscrete, vramMB, computeUnits)
}

func vkVendorName(vendorID uint32) string {
	switch vendorID {
	case 0x10DE:
		return "NVIDIA Corporation"
	case 0x1002, 0x1022:
		return "Advanced Micro Devices, Inc."
	case 0x8086:
		return "Intel(R) Corporation"
	case 0x13B5:
		return "Arm Ltd."
	default:
		return fmt.Sprintf("vendorID=0x%04X", vendorID)
	}
}

func scoreVkDevice(pd vkPhysicalDevice) (int64, uint32) {
	var props C.VkPhysicalDeviceProperties
	C.rawVkGetPhysicalDeviceProperties(pd, &props)
	var qCount C.uint32_t
	C.rawVkGetPhysicalDeviceQueueFamilyProperties(pd, &qCount, nil)
	var qfps *C.VkQueueFamilyProperties
	if qCount > 0 {
		qfps = C.allocQueueFamilyProperties(C.size_t(qCount))
		if qfps != nil {
			defer C.freePtr(unsafe.Pointer(qfps))
			C.rawVkGetPhysicalDeviceQueueFamilyProperties(pd, &qCount, qfps)
		}
	}
	var qi uint32
	found := false
	qfpSlice := unsafe.Slice((*C.VkQueueFamilyProperties)(unsafe.Pointer(qfps)), int(qCount))
	for i, qfp := range qfpSlice {
		if qfp.queueFlags&C.VK_QUEUE_COMPUTE_BIT != 0 {
			qi = uint32(i)
			found = true
			break
		}
	}
	if !found {
		for i, qfp := range qfpSlice {
			if qfp.queueFlags&C.VK_QUEUE_GRAPHICS_BIT != 0 {
				qi = uint32(i)
				break
			}
		}
	}
	var score int64
	if props.deviceType == C.VK_PHYSICAL_DEVICE_TYPE_DISCRETE_GPU {
		score += 1_500_000_000_000
	} else if props.deviceType == C.VK_PHYSICAL_DEVICE_TYPE_INTEGRATED_GPU {
		score += 1_000_000_000_000
	} else if props.deviceType == C.VK_PHYSICAL_DEVICE_TYPE_CPU {
		score += 100_000_000_000
	}
	var memProps C.VkPhysicalDeviceMemoryProperties
	C.rawVkGetPhysicalDeviceMemoryProperties(pd, &memProps)
	var vramMB int64
	for i := C.uint32_t(0); i < memProps.memoryHeapCount; i++ {
		if memProps.memoryHeaps[i].flags&C.VK_MEMORY_HEAP_DEVICE_LOCAL_BIT != 0 {
			vramMB = int64(memProps.memoryHeaps[i].size / (1024 * 1024))
			break
		}
	}
	score += vramMB * 10_000
	if props.limits.maxComputeSharedMemorySize > 0 {
		score += int64(props.limits.maxComputeSharedMemorySize/1024) * 1_000
	}
	name := C.GoString(&props.deviceName[0])
	vendor := strings.ToLower(name)
	if strings.Contains(vendor, "nvidia") || strings.Contains(vendor, "geforce") || strings.Contains(vendor, "rtx") || strings.Contains(vendor, "quadro") {
		score += 5_000_000_000
	} else if strings.Contains(vendor, "amd") || strings.Contains(vendor, "radeon") {
		score += 3_000_000_000
	} else if strings.Contains(vendor, "intel") || strings.Contains(vendor, "arc") {
		score += 1_000_000_000
	}
	return score, qi
}

func vkMakeDevice(pd vkPhysicalDevice, qIdx uint32) (vkDevice, vkQueue, error) {
	var dev vkDevice
	var queue vkQueue
	if res := C.rawVkCreateDeviceSimple(pd, C.uint32_t(qIdx), &dev, &queue); res != C.VK_SUCCESS {
		return dev, queue, fmt.Errorf("vkCreateDevice: %d", int(res))
	}
	return dev, queue, nil
}

func vkMakeCommandPool(dev vkDevice, qIdx uint32) (vkCommandPool, error) {
	var pool vkCommandPool
	if res := C.rawVkCreateCommandPoolSimple(dev, C.uint32_t(qIdx), &pool); res != C.VK_SUCCESS {
		var zero vkCommandPool
		return zero, fmt.Errorf("vkCreateCommandPool: %d", int(res))
	}
	return pool, nil
}

func (v *vulkanBackend) makeFences() error {
	fci := C.VkFenceCreateInfo{
		sType: C.VK_STRUCTURE_TYPE_FENCE_CREATE_INFO,
		flags: C.VK_FENCE_CREATE_SIGNALED_BIT,
	}
	for i := 0; i < ringSize; i++ {
		fence := v.fences[i]
		if res := C.rawVkCreateFence(v.device, &fci, nil, &fence); res != C.VK_SUCCESS {
			return fmt.Errorf("vkCreateFence[%d]: %d", i, int(res))
		}
		v.fences[i] = fence
	}
	return nil
}

func (v *vulkanBackend) allocApplyResources() error {
	ai := C.VkCommandBufferAllocateInfo{
		sType:              C.VK_STRUCTURE_TYPE_COMMAND_BUFFER_ALLOCATE_INFO,
		commandPool:        v.commandPool,
		level:              C.VK_COMMAND_BUFFER_LEVEL_PRIMARY,
		commandBufferCount: C.uint32_t(ringSize),
	}
	var tmp [ringSize]vkCommandBuffer
	if res := C.rawVkCreateCommandBuffer(v.device, &ai, &tmp[0]); res != C.VK_SUCCESS {
		return fmt.Errorf("vkAllocateCommandBuffers(apply): %d", int(res))
	}
	v.applyCmdBufs = tmp
	fci := C.VkFenceCreateInfo{
		sType: C.VK_STRUCTURE_TYPE_FENCE_CREATE_INFO,
		flags: C.VK_FENCE_CREATE_SIGNALED_BIT,
	}
	for i := 0; i < ringSize; i++ {
		fence := v.applyFences[i]
		if res := C.rawVkCreateFence(v.device, &fci, nil, &fence); res != C.VK_SUCCESS {
			for j := 0; j < i; j++ {
				C.rawVkDestroyFence(v.device, v.applyFences[j], nil)
				v.applyFences[j] = nil
			}
			C.rawVkFreeCommandBuffers(v.device, v.commandPool, C.uint32_t(ringSize), &v.applyCmdBufs[0])
			v.applyCmdBufs = [ringSize]vkCommandBuffer{}
			v.applyFences = [ringSize]vkFence{}
			return fmt.Errorf("vkCreateFence(apply[%d]): %d", i, int(res))
		}
		v.applyFences[i] = fence
	}
	return nil
}

func (v *vulkanBackend) allocTransferResources() error {
	ai := C.VkCommandBufferAllocateInfo{
		sType:              C.VK_STRUCTURE_TYPE_COMMAND_BUFFER_ALLOCATE_INFO,
		commandPool:        v.commandPool,
		level:              C.VK_COMMAND_BUFFER_LEVEL_PRIMARY,
		commandBufferCount: 1,
	}
	if res := C.rawVkCreateCommandBuffer(v.device, &ai, &v.transferCmdBuf); res != C.VK_SUCCESS {
		return fmt.Errorf("vkAllocateCommandBuffers(transfer): %d", int(res))
	}
	fci := C.VkFenceCreateInfo{
		sType: C.VK_STRUCTURE_TYPE_FENCE_CREATE_INFO,
		flags: C.VK_FENCE_CREATE_SIGNALED_BIT,
	}
	if res := C.rawVkCreateFence(v.device, &fci, nil, &v.transferFence); res != C.VK_SUCCESS {
		C.rawVkFreeCommandBuffers(v.device, v.commandPool, 1, &v.transferCmdBuf)
		v.transferCmdBuf = nil
		return fmt.Errorf("vkCreateFence(transfer): %d", int(res))
	}
	return nil
}

func (v *vulkanBackend) allocCommandBuffers() error {
	ai := C.VkCommandBufferAllocateInfo{
		sType:              C.VK_STRUCTURE_TYPE_COMMAND_BUFFER_ALLOCATE_INFO,
		commandPool:        v.commandPool,
		level:              C.VK_COMMAND_BUFFER_LEVEL_PRIMARY,
		commandBufferCount: C.uint32_t(ringSize),
	}
	var tmp [ringSize]vkCommandBuffer
	if res := C.rawVkCreateCommandBuffer(v.device, &ai, &tmp[0]); res != C.VK_SUCCESS {
		return fmt.Errorf("vkAllocateCommandBuffers: %d", int(res))
	}
	v.cmdBufs = tmp
	return nil
}

func (v *vulkanBackend) allocGridResources() error {
	ai := C.VkCommandBufferAllocateInfo{
		sType:              C.VK_STRUCTURE_TYPE_COMMAND_BUFFER_ALLOCATE_INFO,
		commandPool:        v.commandPool,
		level:              C.VK_COMMAND_BUFFER_LEVEL_PRIMARY,
		commandBufferCount: C.uint32_t(ringSize),
	}
	var tmp [ringSize]vkCommandBuffer
	if res := C.rawVkCreateCommandBuffer(v.device, &ai, &tmp[0]); res != C.VK_SUCCESS {
		return fmt.Errorf("vkAllocateCommandBuffers(grid): %d", int(res))
	}
	v.gridCmdBufs = tmp
	fci := C.VkFenceCreateInfo{
		sType: C.VK_STRUCTURE_TYPE_FENCE_CREATE_INFO,
		flags: C.VK_FENCE_CREATE_SIGNALED_BIT,
	}
	for i := 0; i < ringSize; i++ {
		fence := v.gridFences[i]
		if res := C.rawVkCreateFence(v.device, &fci, nil, &fence); res != C.VK_SUCCESS {
			for j := 0; j < i; j++ {
				C.rawVkDestroyFence(v.device, v.gridFences[j], nil)
				v.gridFences[j] = nil
			}
			C.rawVkFreeCommandBuffers(v.device, v.commandPool, C.uint32_t(ringSize), &v.gridCmdBufs[0])
			v.gridCmdBufs = [ringSize]vkCommandBuffer{}
			v.gridFences = [ringSize]vkFence{}
			return fmt.Errorf("vkCreateFence(grid[%d]): %d", i, int(res))
		}
		v.gridFences[i] = fence
	}
	return nil
}

func (v *vulkanBackend) beginCommandBuffer(cmdBuf vkCommandBuffer) error {
	if res := C.rawVkResetCommandBufferOne(cmdBuf); res != C.VK_SUCCESS {
		return fmt.Errorf("vkResetCommandBuffer: %d", int(res))
	}
	bi := C.VkCommandBufferBeginInfo{sType: C.VK_STRUCTURE_TYPE_COMMAND_BUFFER_BEGIN_INFO, flags: C.VK_COMMAND_BUFFER_USAGE_ONE_TIME_SUBMIT_BIT}
	if res := C.rawVkBeginCommandBuffer(cmdBuf, &bi); res != C.VK_SUCCESS {
		return fmt.Errorf("vkBeginCommandBuffer: %d", int(res))
	}
	return nil
}

func (v *vulkanBackend) endCommandBuffer(cmdBuf vkCommandBuffer) error {
	if res := C.rawVkEndCommandBuffer(cmdBuf); res != C.VK_SUCCESS {
		return fmt.Errorf("vkEndCommandBuffer: %d", int(res))
	}
	return nil
}

func (v *vulkanBackend) makeComputePipeline(pipe *vkPipeline, layout *vkPipelineLayout, name string, spv []byte) error {
	if len(spv) == 0 || len(spv)%4 != 0 {
		return fmt.Errorf("shader %s: invalid SPIR-V payload", name)
	}
	var shaderModule C.VkShaderModule
	if res := C.rawVkCreateShaderModuleSimple(v.device, (*C.uint32_t)(unsafe.Pointer(&spv[0])), C.size_t(len(spv)), &shaderModule); res != C.VK_SUCCESS {
		return fmt.Errorf("vkCreateShaderModule(%s): %d", name, int(res))
	}
	defer C.rawVkDestroyShaderModule(v.device, shaderModule, nil)

	if res := C.rawVkCreatePipelineLayoutSimple(v.device, v.dsLayout, layout); res != C.VK_SUCCESS {
		return fmt.Errorf("vkCreatePipelineLayout(%s): %d", name, int(res))
	}
	if res := C.rawVkCreateComputePipelineSimple2(v.device, shaderModule, *layout, pipe); res != C.VK_SUCCESS {
		C.rawVkDestroyPipelineLayout(v.device, *layout, nil)
		return fmt.Errorf("vkCreateComputePipelines(%s): %d", name, int(res))
	}
	return nil
}

func (v *vulkanBackend) bindApplyArgs(cmdBuf vkCommandBuffer, slot, xMin, yMin, xMax, yMax int, candidate model.Candidate) error {
	if v.applyPipe == nil || v.applyLayout == nil {
		return fmt.Errorf("apply pipeline not initialized")
	}
	C.rawVkCmdBindPipeline(cmdBuf, C.VK_PIPELINE_BIND_POINT_COMPUTE, v.applyPipe)
	ds := v.applyDs[slot]
	C.rawVkCmdBindDescriptorSets(cmdBuf, C.VK_PIPELINE_BIND_POINT_COMPUTE, v.applyLayout, 0, 1, &ds)
	type applyPC struct {
		Width, Height          int32
		XMin, YMin, XMax, YMax int32
		CX, CY, RXRaw, RYRaw   float32
		ThetaDeg               float32
		CR, CG, CB, CA         float32
		ShapeType              int32
	}
	pc := applyPC{
		Width: int32(v.width), Height: int32(v.height),
		XMin: int32(xMin), YMin: int32(yMin), XMax: int32(xMax), YMax: int32(yMax),
		CX: candidate.X, CY: candidate.Y, RXRaw: candidate.RX, RYRaw: candidate.RY,
		ThetaDeg: candidate.Theta, CR: candidate.R, CG: candidate.G, CB: candidate.B, CA: candidate.A,
		ShapeType: int32(candidate.ShapeType),
	}
	C.rawVkCmdPushConstants(cmdBuf, v.applyLayout, C.VK_SHADER_STAGE_COMPUTE_BIT, 0, C.uint32_t(unsafe.Sizeof(pc)), unsafe.Pointer(&pc))
	return nil
}

func (v *vulkanBackend) bindGridArgs(cmdBuf vkCommandBuffer, slot int) error {
	if v.gridPipe == nil || v.gridLayout == nil {
		return fmt.Errorf("grid pipeline not initialized")
	}
	C.rawVkCmdBindPipeline(cmdBuf, C.VK_PIPELINE_BIND_POINT_COMPUTE, v.gridPipe)
	ds := v.gridDs[slot]
	C.rawVkCmdBindDescriptorSets(cmdBuf, C.VK_PIPELINE_BIND_POINT_COMPUTE, v.gridLayout, 0, 1, &ds)
	pc := [4]int32{int32(v.width), int32(v.height), int32(v.gridW), int32(v.gridH)}
	C.rawVkCmdPushConstants(cmdBuf, v.gridLayout, C.VK_SHADER_STAGE_COMPUTE_BIT, 0, C.uint32_t(unsafe.Sizeof(pc)), unsafe.Pointer(&pc[0]))
	return nil
}

func ensureVulkanLoaded() error { return nil }
